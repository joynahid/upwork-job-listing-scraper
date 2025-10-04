package server

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

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

	if len(opts.ContractorTierCodes) > 0 {
		if job.ContractorTier == nil || !intInSlice(*job.ContractorTier, opts.ContractorTierCodes) {
			return false
		}
	}

	if len(opts.JobTypeCodes) > 0 {
		if job.JobType == nil || !intInSlice(*job.JobType, opts.JobTypeCodes) {
			return false
		}
	}

	if len(opts.DurationLabels) > 0 {
		if strings.TrimSpace(job.DurationLabel) == "" || !stringInSliceFold(job.DurationLabel, opts.DurationLabels) {
			return false
		}
	}

	if len(opts.WorkloadValues) > 0 {
		if !matchesWorkload(job.Workload, opts.WorkloadValues) {
			return false
		}
	}

	if opts.ContractToHire != nil {
		if job.IsContractToHire == nil || *job.IsContractToHire != *opts.ContractToHire {
			return false
		}
	}

	if len(opts.BudgetRanges) > 0 {
		if !matchesBudgetRanges(job, opts.BudgetRanges) {
			return false
		}
	}

	if len(opts.HourlyRanges) > 0 {
		if !matchesHourlyRanges(job, opts.HourlyRanges) {
			return false
		}
	}

	if len(opts.ClientHiresRanges) > 0 {
		if job.Buyer == nil || job.Buyer.TotalJobsWithHires == nil || !intRangeContains(*job.Buyer.TotalJobsWithHires, opts.ClientHiresRanges) {
			return false
		}
	}

	if len(opts.CategoryGroupIDs) > 0 {
		if job.Category == nil || !stringInSliceFold(job.Category.GroupSlug, opts.CategoryGroupIDs) {
			return false
		}
	}

	if len(opts.Proposals) > 0 {
		if job.ProposalsTier == "" || !stringInSliceFold(job.ProposalsTier, opts.Proposals) {
			return false
		}
	}

	if len(opts.LocationRegions) > 0 {
		if !matchesLocationFilters(job, opts.LocationRegions) {
			return false
		}
	}

	if len(opts.Timezones) > 0 {
		if !matchesTimezoneFilters(job, opts.Timezones) {
			return false
		}
	}

	if opts.PreviousClients != "" && !matchesPreviousClients(job, opts.PreviousClients) {
		return false
	}

	return true
}

func intInSlice(value int, list []int) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func stringInSliceFold(value string, list []string) bool {
	for _, item := range list {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}

func matchesWorkload(workload string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	normalizedJob := normalizeToken(workload)
	if normalizedJob == "" {
		return false
	}
	for _, filter := range filters {
		normFilter := normalizeToken(filter)
		if normFilter == "" {
			continue
		}
		if strings.Contains(normalizedJob, normFilter) || strings.Contains(normFilter, normalizedJob) {
			return true
		}
	}
	return false
}

func matchesBudgetRanges(job *JobRecord, ranges []NumericRange) bool {
	if len(ranges) == 0 {
		return true
	}
	if job.Budget == nil || job.Budget.FixedAmount == nil {
		return false
	}
	amount := *job.Budget.FixedAmount
	for _, r := range ranges {
		if r.contains(amount) {
			return true
		}
	}
	return false
}

func matchesHourlyRanges(job *JobRecord, ranges []NumericRange) bool {
	if len(ranges) == 0 {
		return true
	}
	if job.HourlyInfo == nil {
		return false
	}
	var minPtr, maxPtr *float64
	if job.HourlyInfo.Min != nil {
		minCopy := *job.HourlyInfo.Min
		minPtr = &minCopy
	}
	if job.HourlyInfo.Max != nil {
		maxCopy := *job.HourlyInfo.Max
		maxPtr = &maxCopy
	}
	if minPtr == nil && maxPtr == nil {
		return false
	}
	if minPtr == nil {
		minPtr = maxPtr
	}
	if maxPtr == nil {
		maxPtr = minPtr
	}
	minVal := *minPtr
	maxVal := *maxPtr
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	for _, r := range ranges {
		if overlapsFloatRange(minVal, maxVal, r) {
			return true
		}
	}
	return false
}

func overlapsFloatRange(minVal, maxVal float64, r NumericRange) bool {
	if r.Min != nil && maxVal < *r.Min {
		return false
	}
	if r.Max != nil && minVal > *r.Max {
		return false
	}
	return true
}

func intRangeContains(value int, ranges []IntRange) bool {
	for _, r := range ranges {
		if r.contains(value) {
			return true
		}
	}
	return false
}

func matchesLocationFilters(job *JobRecord, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		if matchSingleLocation(job, filter) {
			return true
		}
	}
	return false
}

func matchSingleLocation(job *JobRecord, filter string) bool {
	normalized := strings.ToLower(strings.TrimSpace(filter))
	if normalized == "" {
		return true
	}
	timezones := collectJobTimezones(job)
	countries := collectJobCountries(job)

	switch normalized {
	case "africa":
		if hasTimezonePrefix(timezones, "africa/") {
			return true
		}
	case "europe":
		if hasTimezonePrefix(timezones, "europe/") {
			return true
		}
	case "caribbean":
		for _, country := range countries {
			if _, ok := caribbeanCountrySet[country]; ok {
				return true
			}
		}
	default:
		for _, country := range countries {
			if strings.EqualFold(country, normalized) {
				return true
			}
		}
	}

	for _, tz := range timezones {
		if strings.EqualFold(tz, normalized) || strings.Contains(tz, normalized) {
			return true
		}
	}

	if normalized == "caribbean" {
		for _, tz := range timezones {
			if strings.HasPrefix(tz, "america/") {
				return true
			}
		}
	}

	return false
}

func matchesTimezoneFilters(job *JobRecord, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	timezones := collectJobTimezones(job)
	if len(timezones) == 0 {
		return false
	}
	for _, filter := range filters {
		norm := strings.ToLower(strings.TrimSpace(filter))
		for _, tz := range timezones {
			if strings.EqualFold(tz, norm) {
				return true
			}
		}
	}
	return false
}

func matchesPreviousClients(job *JobRecord, criteria string) bool {
	norm := strings.ToLower(strings.TrimSpace(criteria))
	if norm == "" || norm == "all" {
		return true
	}
	if job.Buyer == nil {
		return false
	}
	total := 0
	if job.Buyer.TotalJobsWithHires != nil {
		total = *job.Buyer.TotalJobsWithHires
	} else if job.Buyer.TotalAssignments != nil {
		total = *job.Buyer.TotalAssignments
	}
	switch norm {
	case "no", "none", "new", "firsttime", "zero":
		return total == 0
	case "yes", "previous", "existing", "returning":
		return total > 0
	default:
		return true
	}
}

func collectJobTimezones(job *JobRecord) []string {
	result := []string{}
	if job.Location != nil && job.Location.Timezone != "" {
		result = append(result, strings.ToLower(job.Location.Timezone))
	}
	if job.Buyer != nil && job.Buyer.Timezone != "" {
		result = append(result, strings.ToLower(job.Buyer.Timezone))
	}
	return result
}

func collectJobCountries(job *JobRecord) []string {
	result := []string{}
	if job.Location != nil && job.Location.Country != "" {
		result = append(result, strings.ToLower(job.Location.Country))
	}
	if job.Buyer != nil && job.Buyer.Country != "" {
		result = append(result, strings.ToLower(job.Buyer.Country))
	}
	return result
}

func hasTimezonePrefix(timezones []string, prefix string) bool {
	for _, tz := range timezones {
		if strings.HasPrefix(tz, prefix) {
			return true
		}
	}
	return false
}

func normalizeToken(value string) string {
	if value == "" {
		return ""
	}
	var builder strings.Builder
	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

var caribbeanCountrySet = map[string]struct{}{
	"ag": {}, "antiguaandbarbuda": {},
	"ai": {}, "anguilla": {},
	"aw": {}, "aruba": {},
	"bs": {}, "bahamas": {},
	"bb": {}, "barbados": {},
	"bz": {}, "belize": {},
	"vg": {}, "britishvirginislands": {},
	"ky": {}, "caymanislands": {},
	"cu": {}, "cuba": {},
	"dm": {}, "dominica": {},
	"do": {}, "dominicanrepublic": {},
	"gd": {}, "grenada": {},
	"gp": {}, "guadeloupe": {},
	"ht": {}, "haiti": {},
	"jm": {}, "jamaica": {},
	"mq": {}, "martinique": {},
	"ms": {}, "montserrat": {},
	"pr": {}, "puertorico": {},
	"kn": {}, "saintkittsandnevis": {},
	"lc": {}, "saintlucia": {},
	"mf": {}, "saintmartin": {},
	"vc": {}, "saintvincentandthegrenadines": {},
	"tt": {}, "trinidadandtobago": {},
	"tc": {}, "turksandcaicosislands": {},
	"vi": {}, "usvirginislands": {},
	"sx": {}, "sintmaarten": {},
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
