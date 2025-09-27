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

	return &JobRecord{
		ID:             id,
		Title:          title,
		Description:    description,
		JobType:        jobType,
		Status:         status,
		ContractorTier: contractorTier,
		Category:       category,
		PostedOn:       postedOn,
		Budget:         budget,
		Buyer:          buyer,
		Tags:           tags,
		URL:            url,
		LastVisitedAt:  lastVisited,
		Skills:         skills,
		HourlyInfo:     hourly,
		ClientActivity: clientActivity,
		Location:       location,
		DurationLabel:  duration,
		Engagement:     engagement,
		IsPrivate:      isPrivate,
		PrivacyReason:  privacyReason,
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

	return &JobSummaryRecord{
		ID:            id,
		Title:         title,
		Description:   description,
		JobType:       jobType,
		DurationLabel: duration,
		Engagement:    engagement,
		Skills:        skills,
		HourlyInfo:    hourly,
		FixedBudget:   fixedBudget,
		WeeklyBudget:  weeklyBudget,
		Client:        client,
		Ciphertext:    cipher,
		URL:           url,
		PublishedOn:   published,
		RenewedOn:     renewed,
		LastVisitedAt: lastVisited,
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
		case SortPostedOn:
			aTime := timeOrZero(a.PublishedOn)
			bTime := timeOrZero(b.PublishedOn)
			if aTime.Equal(bTime) {
				return compareSummaryFallback(a, b, opts.SortAscending)
			}
			if opts.SortAscending {
				if aTime.IsZero() {
					return false
				}
				if bTime.IsZero() {
					return true
				}
				return aTime.Before(bTime)
			}
			if aTime.IsZero() {
				return false
			}
			if bTime.IsZero() {
				return true
			}
			return aTime.After(bTime)
		default:
			aTime := timeOrZero(a.LastVisitedAt)
			bTime := timeOrZero(b.LastVisitedAt)
			if aTime.Equal(bTime) {
				return compareSummaryFallback(a, b, opts.SortAscending)
			}
			if opts.SortAscending {
				if aTime.IsZero() {
					return false
				}
				if bTime.IsZero() {
					return true
				}
				return aTime.Before(bTime)
			}
			if aTime.IsZero() {
				return false
			}
			if bTime.IsZero() {
				return true
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
	}

	if info.PaymentVerified == nil && info.Country == "" && info.City == "" && info.Timezone == "" && info.TotalSpent == nil && info.TotalAssignments == nil && info.TotalJobsWithHires == nil {
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

	if len(opts.Tags) > 0 {
		if len(job.Tags) == 0 {
			return false
		}
		tagSet := make(map[string]struct{}, len(job.Tags))
		for _, tag := range job.Tags {
			tagSet[strings.ToLower(tag)] = struct{}{}
		}
		for _, desired := range opts.Tags {
			if _, ok := tagSet[strings.ToLower(desired)]; !ok {
				return false
			}
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
		case SortPostedOn:
			aTime := timeOrZero(a.PostedOn)
			bTime := timeOrZero(b.PostedOn)
			if aTime.Equal(bTime) {
				return compareFallback(a, b, opts.SortAscending)
			}
			if opts.SortAscending {
				if aTime.IsZero() {
					return false
				}
				if bTime.IsZero() {
					return true
				}
				return aTime.Before(bTime)
			}
			if aTime.IsZero() {
				return false
			}
			if bTime.IsZero() {
				return true
			}
			return aTime.After(bTime)
		default:
			aTime := timeOrZero(a.LastVisitedAt)
			bTime := timeOrZero(b.LastVisitedAt)
			if aTime.Equal(bTime) {
				return compareFallback(a, b, opts.SortAscending)
			}
			if opts.SortAscending {
				if aTime.IsZero() {
					return false
				}
				if bTime.IsZero() {
					return true
				}
				return aTime.Before(bTime)
			}
			if aTime.IsZero() {
				return false
			}
			if bTime.IsZero() {
				return true
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
