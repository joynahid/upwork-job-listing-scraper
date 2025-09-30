package server

import (
	"fmt"
	"sort"
	"strings"

	"cloud.google.com/go/firestore"
)

type jobSource struct {
	data  map[string]interface{}
	buyer map[string]interface{}
}

// transformDocument converts a Firestore snapshot into one or more JobRecords.
func transformDocument(doc *firestore.DocumentSnapshot) ([]JobRecord, error) {
	return debugTransformDocument(doc)
}

// debugTransformDocument exposes the transformation logic for diagnostics.
func DebugTransformDocument(doc *firestore.DocumentSnapshot) ([]JobRecord, error) {
	return debugTransformDocument(doc)
}

func debugTransformDocument(doc *firestore.DocumentSnapshot) ([]JobRecord, error) {
	raw := doc.Data()
	if raw == nil {
		return nil, fmt.Errorf("empty document data")
	}

	stateMap := getMap(raw, "state")
	if stateMap == nil {
		return nil, fmt.Errorf("missing state data")
	}

	jobState := getMap(stateMap, "job")
	isPrivate, privacyReason := detectPrivacy(jobState)

	primaryCandidates := []map[string]interface{}{
		getMap(stateMap, "jobDetails", "job"),
		getMap(stateMap, "job", "job"),
		getMap(stateMap, "job"),
		getMap(raw, "job"),
	}

	var primaryJob map[string]interface{}
	for _, candidate := range primaryCandidates {
		if isValidJobMap(candidate) {
			primaryJob = candidate
			break
		}
	}

	buyerMap := firstNonNilMap(
		getMap(stateMap, "jobDetails", "buyer"),
		getMap(stateMap, "job", "buyer"),
	)

	sources := make([]jobSource, 0)
	if primaryJob != nil {
		sources = append(sources, jobSource{data: primaryJob, buyer: buyerMap})
	}

	if len(sources) == 0 {
		similarJobs := extractMapSlice(stateMap, "job", "errorResponse", "similarJobs")
		for _, jobMap := range similarJobs {
			sources = append(sources, jobSource{data: jobMap})
		}
	}

	if len(sources) == 0 {
		if isPrivate {
			placeholder := buildPrivatePlaceholder(raw, doc.Ref.ID, privacyReason)
			if placeholder != nil {
				return []JobRecord{*placeholder}, nil
			}
		}
		return nil, fmt.Errorf("no usable job payload")
	}

	seen := make(map[string]struct{})
	records := make([]JobRecord, 0, len(sources))
	for _, src := range sources {
		if src.data == nil {
			continue
		}

		rec := buildJobRecord(src.data, src.buyer, raw, doc.Ref.ID, isPrivate, privacyReason)
		if rec == nil || rec.ID == "" {
			continue
		}
		if _, exists := seen[rec.ID]; exists {
			continue
		}
		seen[rec.ID] = struct{}{}
		records = append(records, *rec)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no jobs extracted")
	}

	return records, nil
}

func buildJobRecord(jobMap map[string]interface{}, buyerMap map[string]interface{}, docMap map[string]interface{}, fallbackID string, isPrivate bool, privacyReason string) *JobRecord {
	id := firstNonEmpty(
		getString(jobMap, "uid"),
		fallbackID,
	)

	title, _ := firstString(jobMap,
		[]string{"title"},
		[]string{"jobTitle"},
	)

	description, _ := firstString(jobMap,
		[]string{"description"},
		[]string{"jobDescription"},
	)

	jobType := getIntPointer(jobMap, "type")
	status := getIntPointer(jobMap, "status")
	contractorTier := getIntPointer(jobMap, "contractorTier")

	categoryMap := getMap(jobMap, "category")
	categoryGroupMap := getMap(jobMap, "categoryGroup")
	category := &CategoryInfo{}
	if categoryMap != nil {
		if name := getString(categoryMap, "name"); name != "" {
			category.Name = name
		}
		if slug := getString(categoryMap, "urlSlug"); slug != "" {
			category.Slug = slug
		}
	}
	if categoryGroupMap != nil {
		if name := getString(categoryGroupMap, "name"); name != "" {
			category.Group = name
		}
		if slug := getString(categoryGroupMap, "urlSlug"); slug != "" {
			category.GroupSlug = slug
		}
	}
	if category.Name == "" && category.Slug == "" && category.Group == "" && category.GroupSlug == "" {
		category = nil
	}

	budget, hourly := buildBudget(jobMap)
	buyer := buildBuyer(buyerMap)

	tags, _ := extractStringSlice(jobMap, "annotations", "tags")
	skills := extractSkillLabels(jobMap, "ontologySkills")

	postedOn := firstTime(jobMap,
		[]string{"postedOn"},
		[]string{"publishTime"},
		[]string{"createdOn"},
	)

	lastVisited := firstTime(docMap, []string{"scrape_metadata", "last_visited_at"})

	url := getString(docMap, "url")
	if url == "" {
		if cipher := getString(jobMap, "ciphertext"); cipher != "" {
			url = fmt.Sprintf("https://www.upwork.com/jobs/%s", cipher)
		}
	}

	clientActivity := buildClientActivity(getMap(jobMap, "clientActivity"))
	location := buildJobLocation(jobMap)
	duration := getString(jobMap, "durationLabel")
	engagement := getString(jobMap, "engagement")

	// Extract new fields
	ciphertext := getString(jobMap, "ciphertext")
	workload := getString(jobMap, "workload")
	proposalsTier := getString(jobMap, "proposalsTier")
	tierText := getString(jobMap, "tierText")

	createdOn := firstTime(jobMap, []string{"createdOn"})
	publishTime := firstTime(jobMap, []string{"publishTime"})

	var isContractToHire *bool
	if val, ok := extractBool(jobMap, "contractToHire"); ok {
		isContractToHire = &val
	}

	var numberOfPositions *int
	if val, ok := extractInt(jobMap, "numberOfPositionsToHire"); ok {
		numberOfPositions = &val
	}

	var wasRenewed *bool
	if val, ok := extractBool(jobMap, "wasRenewed"); ok {
		wasRenewed = &val
	}

	var premium *bool
	if val, ok := extractBool(jobMap, "premium"); ok {
		premium = &val
	}

	var hideBudget *bool
	if val, ok := extractBool(jobMap, "hideBudget"); ok {
		hideBudget = &val
	}

	var recno *int64
	if val, ok := extractInt(jobMap, "recno"); ok {
		recno64 := int64(val)
		recno = &recno64
	}

	qualifications := buildQualifications(getMap(jobMap, "qualifications"))
	weeklyRetainerBudget := buildWeeklyRetainerBudget(jobMap)
	occupations := extractOccupations(jobMap)

	return &JobRecord{
		ID:                   id,
		Title:                title,
		Description:          description,
		JobType:              jobType,
		Status:               status,
		ContractorTier:       contractorTier,
		Category:             category,
		PostedOn:             postedOn,
		CreatedOn:            createdOn,
		PublishTime:          publishTime,
		Budget:               budget,
		Buyer:                buyer,
		Tags:                 tags,
		URL:                  url,
		LastVisitedAt:        lastVisited,
		Skills:               skills,
		HourlyInfo:           hourly,
		ClientActivity:       clientActivity,
		Location:             location,
		DurationLabel:        duration,
		Engagement:           engagement,
		IsPrivate:            isPrivate,
		PrivacyReason:        privacyReason,
		Ciphertext:           ciphertext,
		Workload:             workload,
		IsContractToHire:     isContractToHire,
		NumberOfPositions:    numberOfPositions,
		WasRenewed:           wasRenewed,
		Premium:              premium,
		HideBudget:           hideBudget,
		ProposalsTier:        proposalsTier,
		TierText:             tierText,
		Qualifications:       qualifications,
		WeeklyRetainerBudget: weeklyRetainerBudget,
		Occupations:          occupations,
		Recno:                recno,
	}
}

func buildBudget(job map[string]interface{}) (*BudgetInfo, *HourlyBudget) {
	var fixedAmount *float64
	var currency string

	if budgetMap := getMap(job, "budget"); budgetMap != nil {
		if amount, ok := extractFloat(budgetMap, "amount"); ok {
			v := amount
			fixedAmount = &v
		}
		if curr := getString(budgetMap, "currencyCode"); curr != "" {
			currency = curr
		}
	}

	if amountMap := getMap(job, "amount"); amountMap != nil {
		if amount, ok := extractFloat(amountMap, "amount"); ok {
			v := amount
			fixedAmount = &v
		}
		if curr := getString(amountMap, "currencyCode"); curr != "" {
			currency = curr
		}
	}

	var hourlyMin *float64
	var hourlyMax *float64
	if min, ok := extractFloat(job, "hourlyBudgetMin"); ok {
		v := min
		hourlyMin = &v
	}
	if max, ok := extractFloat(job, "hourlyBudgetMax"); ok {
		v := max
		hourlyMax = &v
	}

	hourlyCurrency := getString(job, "hourlyBudgetCurrencyCode")
	if hourlyCurrency == "" {
		hourlyCurrency = currency
	}

	var budget *BudgetInfo
	if fixedAmount != nil || currency != "" {
		budget = &BudgetInfo{
			FixedAmount: fixedAmount,
			Currency:    currency,
		}
	}

	var hourly *HourlyBudget
	if hourlyMin != nil || hourlyMax != nil || hourlyCurrency != "" {
		hourly = &HourlyBudget{
			Min:      hourlyMin,
			Max:      hourlyMax,
			Currency: hourlyCurrency,
		}
	}

	return budget, hourly
}

func buildPrivatePlaceholder(docMap map[string]interface{}, fallbackID string, reason string) *JobRecord {
	lastVisited := firstTime(docMap, []string{"scrape_metadata", "last_visited_at"})
	url := getString(docMap, "url")

	return &JobRecord{
		ID:            fallbackID,
		URL:           url,
		LastVisitedAt: lastVisited,
		IsPrivate:     true,
		PrivacyReason: reason,
	}
}

func detectPrivacy(jobState map[string]interface{}) (bool, string) {
	if jobState == nil {
		return false, ""
	}

	if errResp := getMap(jobState, "errorResponse"); errResp != nil {
		if status, ok := extractInt(errResp, "status"); ok && status == 403 {
			reason := strings.TrimSpace(getString(errResp, "text"))
			if reason == "" || strings.HasPrefix(reason, "{") {
				reason = "This job is private or restricted (403)."
			}
			return true, reason
		}
	}

	return false, ""
}

func transformJobListDocument(doc *firestore.DocumentSnapshot) (*JobSummaryRecord, error) {
	raw := doc.Data()
	if raw == nil {
		return nil, fmt.Errorf("empty document data")
	}

	id := firstNonEmpty(getString(raw, "uid"), doc.Ref.ID)
	if id == "" {
		return nil, fmt.Errorf("job list document missing uid")
	}

	title := getString(raw, "title")
	description := getString(raw, "description")
	jobType := getIntPointer(raw, "type")
	duration := getString(raw, "durationLabel")
	engagement := getString(raw, "engagement")
	skills := extractSkillLabels(raw, "attrs")

	fixedBudget, hourly := buildBudget(raw)
	weeklyBudget := budgetFromAmount(getMap(raw, "weeklyBudget"))
	client := buildJobSummaryClient(getMap(raw, "client"))

	cipher := getString(raw, "ciphertext")
	url := getString(raw, "url")
	if url == "" && cipher != "" {
		url = fmt.Sprintf("https://www.upwork.com/jobs/%s", cipher)
	}

	published := firstTime(raw,
		[]string{"publishedOn"},
		[]string{"createdOn"},
	)
	renewed := firstTime(raw, []string{"renewedOn"})
	lastVisited := firstTime(raw, []string{"scrape_metadata", "last_visited_at"})

	// Extract new fields
	workload := getString(raw, "workload")
	proposalsTier := getString(raw, "proposalsTier")

	var isContractToHire *bool
	if val, ok := extractBool(raw, "contractToHire"); ok {
		isContractToHire = &val
	}

	var numberOfPositions *int
	if val, ok := extractInt(raw, "numberOfPositionsToHire"); ok {
		numberOfPositions = &val
	}

	var wasRenewed *bool
	if val, ok := extractBool(raw, "wasRenewed"); ok {
		wasRenewed = &val
	}

	var premium *bool
	if val, ok := extractBool(raw, "premium"); ok {
		premium = &val
	}

	var hideBudget *bool
	if val, ok := extractBool(raw, "hideBudget"); ok {
		hideBudget = &val
	}

	var recno *int64
	if val, ok := extractInt(raw, "recno"); ok {
		recno64 := int64(val)
		recno = &recno64
	}

	qualifications := buildQualifications(getMap(raw, "qualifications"))
	weeklyRetainerBudget := buildWeeklyRetainerBudget(raw)
	occupations := extractOccupations(raw)

	return &JobSummaryRecord{
		ID:                   id,
		Title:                title,
		Description:          description,
		JobType:              jobType,
		DurationLabel:        duration,
		Engagement:           engagement,
		Skills:               skills,
		HourlyInfo:           hourly,
		FixedBudget:          fixedBudget,
		WeeklyBudget:         weeklyBudget,
		Client:               client,
		Ciphertext:           cipher,
		URL:                  url,
		PublishedOn:          published,
		RenewedOn:            renewed,
		LastVisitedAt:        lastVisited,
		Workload:             workload,
		IsContractToHire:     isContractToHire,
		NumberOfPositions:    numberOfPositions,
		WasRenewed:           wasRenewed,
		Premium:              premium,
		HideBudget:           hideBudget,
		ProposalsTier:        proposalsTier,
		Qualifications:       qualifications,
		WeeklyRetainerBudget: weeklyRetainerBudget,
		Occupations:          occupations,
		Recno:                recno,
	}, nil
}

func buildJobSummaryClient(client map[string]interface{}) *JobSummaryClient {
	if client == nil {
		return nil
	}

	result := &JobSummaryClient{}
	if val, ok := extractBool(client, "isPaymentVerified"); ok {
		result.PaymentVerified = &val
	}
	if loc := getMap(client, "location"); loc != nil {
		if country := getString(loc, "country"); country != "" {
			result.Country = strings.ToUpper(country)
		}
	}

	if result.PaymentVerified == nil && result.Country == "" {
		return nil
	}
	return result
}

func budgetFromAmount(m map[string]interface{}) *BudgetInfo {
	if m == nil {
		return nil
	}
	if amount, ok := extractFloat(m, "amount"); ok {
		value := amount
		return &BudgetInfo{FixedAmount: &value}
	}
	return nil
}

func applyJobListFilters(job *JobSummaryRecord, opts JobListFilterOptions) bool {
	if job == nil {
		return false
	}

	if opts.PaymentVerified != nil {
		if job.Client == nil || job.Client.PaymentVerified == nil || *job.Client.PaymentVerified != *opts.PaymentVerified {
			return false
		}
	}

	if opts.Country != "" {
		if job.Client == nil || !strings.EqualFold(job.Client.Country, opts.Country) {
			return false
		}
	}

	if len(opts.Skills) > 0 {
		if len(job.Skills) == 0 {
			return false
		}
		set := make(map[string]struct{}, len(job.Skills))
		for _, skill := range job.Skills {
			set[strings.ToLower(skill)] = struct{}{}
		}
		for _, required := range opts.Skills {
			if _, ok := set[strings.ToLower(required)]; !ok {
				return false
			}
		}
	}

	if opts.JobType != nil {
		if job.JobType == nil || *job.JobType != *opts.JobType {
			return false
		}
	}

	if opts.Duration != "" {
		if job.DurationLabel == "" || !strings.EqualFold(job.DurationLabel, opts.Duration) {
			return false
		}
	}

	if opts.MinHourly != nil {
		if job.HourlyInfo == nil || job.HourlyInfo.Min == nil || *job.HourlyInfo.Min < *opts.MinHourly {
			return false
		}
	}

	if opts.MaxHourly != nil {
		if job.HourlyInfo == nil || job.HourlyInfo.Max == nil || *job.HourlyInfo.Max > *opts.MaxHourly {
			return false
		}
	}

	if opts.BudgetMin != nil {
		if job.FixedBudget == nil || job.FixedBudget.FixedAmount == nil || *job.FixedBudget.FixedAmount < *opts.BudgetMin {
			return false
		}
	}

	if opts.BudgetMax != nil {
		if job.FixedBudget == nil || job.FixedBudget.FixedAmount == nil || *job.FixedBudget.FixedAmount > *opts.BudgetMax {
			return false
		}
	}

	if opts.Search != "" {
		needle := strings.ToLower(opts.Search)
		if !strings.Contains(strings.ToLower(job.Title), needle) && !strings.Contains(strings.ToLower(job.Description), needle) {
			return false
		}
	}

	return true
}

func sortJobSummaries(jobs []JobSummaryRecord, opts JobListFilterOptions) {
	if len(jobs) <= 1 {
		return
	}

	sort.SliceStable(jobs, func(i, j int) bool {
		a := jobs[i]
		b := jobs[j]

		switch opts.SortField {
		case SortPublishTime:
			aTime := timeOrZero(a.PublishedOn)
			bTime := timeOrZero(b.PublishedOn)

			// Handle nil/zero times - always sort them to the end regardless of direction
			aZero := aTime.IsZero()
			bZero := bTime.IsZero()

			if aZero && bZero {
				return compareSummaryFallback(a, b, opts.SortAscending)
			}
			if aZero {
				return false // a goes to the end
			}
			if bZero {
				return true // b goes to the end, a comes first
			}

			if aTime.Equal(bTime) {
				return compareSummaryFallback(a, b, opts.SortAscending)
			}

			if opts.SortAscending {
				return aTime.Before(bTime)
			}
			return aTime.After(bTime)
		default:
			aTime := timeOrZero(a.LastVisitedAt)
			bTime := timeOrZero(b.LastVisitedAt)

			// Handle nil/zero times - always sort them to the end regardless of direction
			aZero := aTime.IsZero()
			bZero := bTime.IsZero()

			if aZero && bZero {
				return compareSummaryFallback(a, b, opts.SortAscending)
			}
			if aZero {
				return false // a goes to the end
			}
			if bZero {
				return true // b goes to the end, a comes first
			}

			if aTime.Equal(bTime) {
				return compareSummaryFallback(a, b, opts.SortAscending)
			}

			if opts.SortAscending {
				return aTime.Before(bTime)
			}
			return aTime.After(bTime)
		}
	})
}

func compareSummaryFallback(a JobSummaryRecord, b JobSummaryRecord, ascending bool) bool {
	if ascending {
		return strings.Compare(a.ID, b.ID) < 0
	}
	return strings.Compare(a.ID, b.ID) > 0
}

func buildBuyer(buyer map[string]interface{}) *BuyerInfo {
	if buyer == nil {
		return nil
	}

	info := &BuyerInfo{}

	if val, ok := extractBool(buyer, "isPaymentMethodVerified"); ok {
		info.PaymentVerified = &val
	}

	if loc := getMap(buyer, "location"); loc != nil {
		if country := getString(loc, "country"); country != "" {
			info.Country = strings.ToUpper(country)
		}
		if city := getString(loc, "city"); city != "" {
			info.City = strings.TrimSpace(city)
		}
		if tz := getString(loc, "countryTimezone"); tz != "" {
			info.Timezone = tz
		}
	}

	if stats := getMap(buyer, "stats"); stats != nil {
		if spentMap := getMap(stats, "totalCharges"); spentMap != nil {
			if amount, ok := extractFloat(spentMap, "amount"); ok {
				info.TotalSpent = ptrFloat(amount)
			}
		}
		if val, ok := extractInt(stats, "totalAssignments"); ok {
			info.TotalAssignments = &val
		}
		if val, ok := extractInt(stats, "totalJobsWithHires"); ok {
			info.TotalJobsWithHires = &val
		}
		if val, ok := extractInt(stats, "activeAssignmentsCount"); ok {
			info.ActiveAssignments = &val
		}
		if val, ok := extractInt(stats, "feedbackCount"); ok {
			info.FeedbackCount = &val
		}
		if val, ok := extractFloat(stats, "hoursCount"); ok {
			info.TotalHours = ptrFloat(val)
		}
		if val, ok := extractFloat(stats, "score"); ok {
			info.Score = ptrFloat(val)
		}
	}

	if company := getMap(buyer, "company"); company != nil {
		if industry := getString(company, "industry"); industry != "" {
			info.CompanyIndustry = industry
		}
		if size, ok := extractInt(company, "size"); ok {
			info.CompanySize = &size
		}
		if contractDate := getString(company, "contractDate"); contractDate != "" {
			if parsed := firstTime(company, []string{"contractDate"}); parsed != nil {
				info.ContractDate = parsed
			}
		}
	}

	if jobs := getMap(buyer, "jobs"); jobs != nil {
		if openCount, ok := extractInt(jobs, "openCount"); ok {
			info.OpenJobsCount = &openCount
		}
	}

	if info.PaymentVerified == nil && info.Country == "" && info.City == "" && info.Timezone == "" &&
		info.TotalSpent == nil && info.TotalAssignments == nil && info.TotalJobsWithHires == nil &&
		info.ActiveAssignments == nil && info.FeedbackCount == nil && info.TotalHours == nil &&
		info.Score == nil && info.CompanyIndustry == "" && info.CompanySize == nil &&
		info.ContractDate == nil && info.OpenJobsCount == nil {
		return nil
	}

	return info
}

func buildClientActivity(activity map[string]interface{}) *ClientActivity {
	if activity == nil {
		return nil
	}

	result := &ClientActivity{}

	if v, ok := extractInt(activity, "totalApplicants"); ok {
		result.TotalApplicants = &v
	}
	if v, ok := extractInt(activity, "totalHired"); ok {
		result.TotalHired = &v
	}
	if v, ok := extractInt(activity, "totalInvitedToInterview"); ok {
		result.TotalInvitedToInterview = &v
	}
	if v, ok := extractInt(activity, "unansweredInvites"); ok {
		result.UnansweredInvites = &v
	}
	if v, ok := extractInt(activity, "invitationsSent"); ok {
		result.InvitationsSent = &v
	}
	if last := getString(activity, "lastBuyerActivity"); last != "" {
		result.LastBuyerActivity = last
	}

	if result.TotalApplicants == nil && result.TotalHired == nil && result.TotalInvitedToInterview == nil && result.UnansweredInvites == nil && result.InvitationsSent == nil && result.LastBuyerActivity == "" {
		return nil
	}

	return result
}

func buildJobLocation(job map[string]interface{}) *JobLocation {
	loc := getMap(job, "jobLocation")
	if loc == nil {
		loc = getMap(job, "location")
	}
	if loc == nil {
		return nil
	}

	location := &JobLocation{}
	if country := getString(loc, "country"); country != "" {
		location.Country = strings.ToUpper(country)
	}
	if city := getString(loc, "city"); city != "" {
		location.City = strings.TrimSpace(city)
	}
	if tz := getString(loc, "timezone"); tz != "" {
		location.Timezone = tz
	} else if tz := getString(loc, "countryTimezone"); tz != "" {
		location.Timezone = tz
	}

	if location.Country == "" && location.City == "" && location.Timezone == "" {
		return nil
	}

	return location
}

func extractSkillLabels(root map[string]interface{}, keys ...string) []string {
	items := extractMapSlice(root, keys...)
	if len(items) == 0 {
		return nil
	}

	labels := make([]string, 0, len(items))
	for _, item := range items {
		if label := getString(item, "prefLabel"); label != "" {
			labels = append(labels, label)
		}
	}

	if len(labels) == 0 {
		return nil
	}

	return labels
}

func buildQualifications(quals map[string]interface{}) *JobQualifications {
	if quals == nil {
		return nil
	}

	result := &JobQualifications{}
	hasAny := false

	if val, ok := extractInt(quals, "minJobSuccessScore"); ok {
		result.MinJobSuccessScore = &val
		hasAny = true
	}
	if val, ok := extractInt(quals, "minOdeskHours"); ok {
		result.MinOdeskHours = &val
		hasAny = true
	}
	if val, ok := extractInt(quals, "prefEnglishSkill"); ok {
		result.PrefEnglishSkill = &val
		hasAny = true
	}
	if val, ok := extractBool(quals, "risingTalent"); ok {
		result.RisingTalent = &val
		hasAny = true
	}
	if val, ok := extractBool(quals, "shouldHavePortfolio"); ok {
		result.ShouldHavePortfolio = &val
		hasAny = true
	}
	if val, ok := extractFloat(quals, "minHoursWeek"); ok {
		result.MinHoursWeek = &val
		hasAny = true
	}

	if !hasAny {
		return nil
	}
	return result
}

func buildWeeklyRetainerBudget(job map[string]interface{}) *BudgetInfo {
	retainer := getMap(job, "weeklyRetainerBudget")
	if retainer == nil {
		return nil
	}

	if amount, ok := extractFloat(retainer, "amount"); ok {
		return &BudgetInfo{
			FixedAmount: &amount,
			Currency:    getString(retainer, "currencyCode"),
		}
	}
	return nil
}

func extractOccupations(job map[string]interface{}) []string {
	occList := extractMapSlice(job, "occupations", "occupation")
	if len(occList) == 0 {
		return nil
	}

	result := make([]string, 0, len(occList))
	for _, occ := range occList {
		if name := getString(occ, "prefLabel"); name != "" {
			result = append(result, name)
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func applyFilters(job *JobRecord, opts FilterOptions) bool {
	if job == nil {
		return false
	}

	if opts.PaymentVerified != nil {
		if job.Buyer == nil || job.Buyer.PaymentVerified == nil || *job.Buyer.PaymentVerified != *opts.PaymentVerified {
			return false
		}
	}

	if opts.CategorySlug != "" {
		if job.Category == nil || !strings.EqualFold(job.Category.Slug, opts.CategorySlug) {
			return false
		}
	}

	if opts.CategoryGroupSlug != "" {
		if job.Category == nil || !strings.EqualFold(job.Category.GroupSlug, opts.CategoryGroupSlug) {
			return false
		}
	}

	if opts.Status != nil {
		if job.Status == nil || *job.Status != *opts.Status {
			return false
		}
	}

	if opts.JobType != nil {
		if job.JobType == nil || *job.JobType != *opts.JobType {
			return false
		}
	}

	if opts.ContractorTier != nil {
		if job.ContractorTier == nil || *job.ContractorTier != *opts.ContractorTier {
			return false
		}
	}

	if opts.Country != "" {
		if job.Buyer == nil || !strings.EqualFold(job.Buyer.Country, opts.Country) {
			return false
		}
	}

	if opts.DurationLabel != "" {
		if strings.TrimSpace(job.DurationLabel) == "" || !strings.EqualFold(job.DurationLabel, opts.DurationLabel) {
			return false
		}
	}

	if opts.Engagement != "" {
		if strings.TrimSpace(job.Engagement) == "" || !strings.EqualFold(job.Engagement, opts.Engagement) {
			return false
		}
	}

	if opts.BudgetMin != nil {
		if job.Budget == nil || job.Budget.FixedAmount == nil || *job.Budget.FixedAmount < *opts.BudgetMin {
			return false
		}
	}

	if opts.BudgetMax != nil {
		if job.Budget == nil || job.Budget.FixedAmount == nil || *job.Budget.FixedAmount > *opts.BudgetMax {
			return false
		}
	}

	if opts.HourlyMin != nil {
		if job.HourlyInfo == nil {
			return false
		}
		candidate := job.HourlyInfo.Min
		if candidate == nil {
			candidate = job.HourlyInfo.Max
		}
		if candidate == nil || *candidate < *opts.HourlyMin {
			return false
		}
	}

	if opts.HourlyMax != nil {
		if job.HourlyInfo == nil {
			return false
		}
		candidate := job.HourlyInfo.Max
		if candidate == nil {
			candidate = job.HourlyInfo.Min
		}
		if candidate == nil || *candidate > *opts.HourlyMax {
			return false
		}
	}

	if opts.PostedAfter != nil {
		if job.PostedOn == nil || job.PostedOn.Before(*opts.PostedAfter) {
			return false
		}
	}

	if opts.PostedBefore != nil {
		if job.PostedOn == nil || job.PostedOn.After(*opts.PostedBefore) {
			return false
		}
	}

	if opts.LastVisitedAfter != nil {
		if job.LastVisitedAt == nil || job.LastVisitedAt.Before(*opts.LastVisitedAfter) {
			return false
		}
	}

	if len(opts.Tags) > 0 {
		if len(job.Tags) == 0 && len(job.Skills) == 0 {
			return false
		}
		keywordSet := make(map[string]struct{}, len(job.Tags)+len(job.Skills))
		for _, tag := range job.Tags {
			keywordSet[strings.ToLower(tag)] = struct{}{}
		}
		for _, skill := range job.Skills {
			keywordSet[strings.ToLower(skill)] = struct{}{}
		}
		matched := false
		for _, desired := range opts.Tags {
			if _, ok := keywordSet[strings.ToLower(desired)]; ok {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if opts.BuyerTotalSpentMin != nil {
		if job.Buyer == nil || job.Buyer.TotalSpent == nil || *job.Buyer.TotalSpent < *opts.BuyerTotalSpentMin {
			return false
		}
	}

	if opts.BuyerTotalSpentMax != nil {
		if job.Buyer == nil || job.Buyer.TotalSpent == nil || *job.Buyer.TotalSpent > *opts.BuyerTotalSpentMax {
			return false
		}
	}

	if opts.BuyerTotalAssignmentsMin != nil {
		if job.Buyer == nil || job.Buyer.TotalAssignments == nil || *job.Buyer.TotalAssignments < *opts.BuyerTotalAssignmentsMin {
			return false
		}
	}

	if opts.BuyerTotalAssignmentsMax != nil {
		if job.Buyer == nil || job.Buyer.TotalAssignments == nil || *job.Buyer.TotalAssignments > *opts.BuyerTotalAssignmentsMax {
			return false
		}
	}

	if opts.BuyerTotalJobsWithHiresMin != nil {
		if job.Buyer == nil || job.Buyer.TotalJobsWithHires == nil || *job.Buyer.TotalJobsWithHires < *opts.BuyerTotalJobsWithHiresMin {
			return false
		}
	}

	if opts.BuyerTotalJobsWithHiresMax != nil {
		if job.Buyer == nil || job.Buyer.TotalJobsWithHires == nil || *job.Buyer.TotalJobsWithHires > *opts.BuyerTotalJobsWithHiresMax {
			return false
		}
	}

	// New filters
	if opts.Workload != "" {
		if job.Workload == "" || !strings.EqualFold(job.Workload, opts.Workload) {
			return false
		}
	}

	if opts.IsContractToHire != nil {
		if job.IsContractToHire == nil || *job.IsContractToHire != *opts.IsContractToHire {
			return false
		}
	}

	if opts.NumberOfPositionsMin != nil {
		if job.NumberOfPositions == nil || *job.NumberOfPositions < *opts.NumberOfPositionsMin {
			return false
		}
	}

	if opts.NumberOfPositionsMax != nil {
		if job.NumberOfPositions == nil || *job.NumberOfPositions > *opts.NumberOfPositionsMax {
			return false
		}
	}

	if opts.WasRenewed != nil {
		if job.WasRenewed == nil || *job.WasRenewed != *opts.WasRenewed {
			return false
		}
	}

	if opts.Premium != nil {
		if job.Premium == nil || *job.Premium != *opts.Premium {
			return false
		}
	}

	if opts.HideBudget != nil {
		if job.HideBudget == nil || *job.HideBudget != *opts.HideBudget {
			return false
		}
	}

	if opts.ProposalsTier != "" {
		if job.ProposalsTier == "" || !strings.EqualFold(job.ProposalsTier, opts.ProposalsTier) {
			return false
		}
	}

	// Qualification filters
	if job.Qualifications != nil {
		if opts.MinJobSuccessScore != nil {
			if job.Qualifications.MinJobSuccessScore == nil || *job.Qualifications.MinJobSuccessScore > *opts.MinJobSuccessScore {
				return false
			}
		}

		if opts.MinOdeskHours != nil {
			if job.Qualifications.MinOdeskHours == nil || *job.Qualifications.MinOdeskHours > *opts.MinOdeskHours {
				return false
			}
		}

		if opts.PrefEnglishSkill != nil {
			if job.Qualifications.PrefEnglishSkill == nil || *job.Qualifications.PrefEnglishSkill > *opts.PrefEnglishSkill {
				return false
			}
		}

		if opts.RisingTalent != nil {
			if job.Qualifications.RisingTalent == nil || *job.Qualifications.RisingTalent != *opts.RisingTalent {
				return false
			}
		}

		if opts.ShouldHavePortfolio != nil {
			if job.Qualifications.ShouldHavePortfolio == nil || *job.Qualifications.ShouldHavePortfolio != *opts.ShouldHavePortfolio {
				return false
			}
		}

		if opts.MinHoursWeek != nil {
			if job.Qualifications.MinHoursWeek == nil || *job.Qualifications.MinHoursWeek > *opts.MinHoursWeek {
				return false
			}
		}
	} else {
		// If qualifications is nil but filter requires qualification checks, exclude the job
		if opts.MinJobSuccessScore != nil || opts.MinOdeskHours != nil || opts.PrefEnglishSkill != nil ||
			opts.RisingTalent != nil || opts.ShouldHavePortfolio != nil || opts.MinHoursWeek != nil {
			return false
		}
	}

	return true
}

func sortJobs(jobs []JobRecord, opts FilterOptions) {
	if len(jobs) <= 1 {
		return
	}

	sort.SliceStable(jobs, func(i, j int) bool {
		a := jobs[i]
		b := jobs[j]

		switch opts.SortField {
		case SortPublishTime:
			aTime := timeOrZero(a.PublishTime)
			bTime := timeOrZero(b.PublishTime)

			// Handle nil/zero times - always sort them to the end regardless of direction
			aZero := aTime.IsZero()
			bZero := bTime.IsZero()

			if aZero && bZero {
				return compareFallback(a, b, opts.SortAscending)
			}
			if aZero {
				return false // a goes to the end
			}
			if bZero {
				return true // b goes to the end, a comes first
			}

			if aTime.Equal(bTime) {
				return compareFallback(a, b, opts.SortAscending)
			}

			if opts.SortAscending {
				return aTime.Before(bTime)
			}
			return aTime.After(bTime)
		case SortBudget:
			aValue, aOK := budgetMetric(a)
			bValue, bOK := budgetMetric(b)
			if !aOK && !bOK {
				return compareFallback(a, b, opts.SortAscending)
			}
			if !aOK {
				return false
			}
			if !bOK {
				return true
			}
			if aValue == bValue {
				return compareFallback(a, b, opts.SortAscending)
			}
			if opts.SortAscending {
				return aValue < bValue
			}
			return aValue > bValue
		default:
			aTime := timeOrZero(a.LastVisitedAt)
			bTime := timeOrZero(b.LastVisitedAt)

			// Handle nil/zero times - always sort them to the end regardless of direction
			aZero := aTime.IsZero()
			bZero := bTime.IsZero()

			if aZero && bZero {
				return compareFallback(a, b, opts.SortAscending)
			}
			if aZero {
				return false // a goes to the end
			}
			if bZero {
				return true // b goes to the end, a comes first
			}

			if aTime.Equal(bTime) {
				return compareFallback(a, b, opts.SortAscending)
			}

			if opts.SortAscending {
				return aTime.Before(bTime)
			}
			return aTime.After(bTime)
		}
	})
}

func compareFallback(a JobRecord, b JobRecord, ascending bool) bool {
	if ascending {
		return strings.Compare(a.ID, b.ID) < 0
	}
	return strings.Compare(a.ID, b.ID) > 0
}

func budgetMetric(job JobRecord) (float64, bool) {
	if job.Budget != nil && job.Budget.FixedAmount != nil {
		return *job.Budget.FixedAmount, true
	}
	if job.HourlyInfo != nil {
		if job.HourlyInfo.Max != nil {
			return *job.HourlyInfo.Max, true
		}
		if job.HourlyInfo.Min != nil {
			return *job.HourlyInfo.Min, true
		}
	}
	return 0, false
}
