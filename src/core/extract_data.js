{
    // Get job ID from URL
    const job_id = window.location.pathname.match(/_~(\\d+)/)?.[1] || null;

    // Get title - it's always in h1
    const title = document.querySelector('h1')?.textContent?.trim() || null;

    // Get description - use the specific XPath provided
    let description = null;
    try {
        const descriptionElement = document.evaluate(
            '//*[@id="main"]/div/div/div[1]/div/div/section[1]',
            document,
            null,
            XPathResult.FIRST_ORDERED_NODE_TYPE,
            null
        ).singleNodeValue;

        if (descriptionElement) {
            description = descriptionElement.textContent?.trim() || null;
        }
    } catch (e) { }

    let hourly_rate = null;
    let budget = null;
    let job_type = null;
    let experience_level = null;
    let duration = null;
    let work_hours = null;
    let location_type = null;
    let project_type = null;

    // Enhanced job features extraction - add this to your extract_data.js
    try {
        const featuresSection = document.evaluate(
            '//*[@id="main"]/div/div/div[1]/div/div/section[2]',
            document,
            null,
            XPathResult.FIRST_ORDERED_NODE_TYPE,
            null
        ).singleNodeValue;

        if (featuresSection) {
            const featureItems = featuresSection.querySelectorAll('li');

            featureItems.forEach(item => {
                const iconDiv = item.querySelector('[data-cy]');
                if (!iconDiv) return;

                const featureType = iconDiv.getAttribute('data-cy');

                switch (featureType) {
                    case 'fixed-price':
                        // Fixed-price budget
                        const budgetElement = item.querySelector('[data-test="BudgetAmount"] strong');
                        if (budgetElement) {
                            budget = budgetElement.textContent.trim();
                            job_type = 'Fixed-price';
                        }
                        break;

                    case 'clock-timelog':
                        // Hourly rate range
                        const rateElements = item.querySelectorAll('[data-test="BudgetAmount"] strong');
                        if (rateElements.length >= 2) {
                            const minRate = rateElements[0].textContent.trim();
                            const maxRate = rateElements[1].textContent.trim();
                            hourly_rate = `${minRate} - ${maxRate}/hr`;
                            job_type = 'Hourly';
                        } else if (rateElements.length === 1) {
                            hourly_rate = `${rateElements[0].textContent.trim()}/hr`;
                            job_type = 'Hourly';
                        }
                        break;

                    case 'clock-hourly':
                        // Work hours per week
                        const hoursElement = item.querySelector('strong');
                        if (hoursElement) {
                            work_hours = hoursElement.textContent.trim();
                        }
                        break;

                    case 'duration2':
                        // Project duration
                        const durationElement = item.querySelector('strong');
                        if (durationElement) {
                            duration = durationElement.textContent.trim();
                        }
                        break;

                    case 'expertise':
                        // Experience level
                        const experienceElement = item.querySelector('strong');
                        if (experienceElement) {
                            experience_level = experienceElement.textContent.trim();
                        }
                        break;

                    case 'local':
                        // Location type
                        const locationElement = item.querySelector('strong');
                        if (locationElement) {
                            location_type = locationElement.textContent.trim();
                        }
                        break;

                    case 'briefcase-outlined':
                        // Project type
                        const projectElement = item.querySelector('strong');
                        if (projectElement) {
                            project_type = projectElement.textContent.trim();
                        }
                        break;
                }
            });
        }
    } catch (e) {
        console.log('Error extracting features:', e);
    }

    // Get posted time - use the specific XPath provided
    let posted_date = null;
    try {
        const postedElement = document.evaluate(
            '//*[@id="main"]/div/div/div[1]/div/div/header/div[1]/div',
            document,
            null,
            XPathResult.FIRST_ORDERED_NODE_TYPE,
            null
        ).singleNodeValue;

        if (postedElement) {
            posted_date = postedElement.textContent?.trim() || null;
        }
    } catch (e) { }

    // Get activity section data - extract all title-value pairs
    let proposals_count = null;
    let last_viewed_by_client = null;
    let hires = null;
    let interviewing = null;
    let invites_sent = null;
    let unanswered_invites = null;

    try {
        const activitySection = document.evaluate(
            '//*[@id="main"]/div/div/div[1]/div/div/section[4]',
            document,
            null,
            XPathResult.FIRST_ORDERED_NODE_TYPE,
            null
        ).singleNodeValue;

        if (activitySection) {
            // Get all activity items
            const activityItems = activitySection.querySelectorAll('.ca-item');

            activityItems.forEach(item => {
                const titleSpan = item.querySelector('.title');
                const valueSpan = item.querySelector('.value');

                if (titleSpan && valueSpan) {
                    const title = titleSpan.textContent.trim().replace(':', '').toLowerCase();
                    const value = valueSpan.textContent.trim();

                    // Map titles to variables
                    if (title.includes('proposals')) {
                        proposals_count = value;
                    } else if (title.includes('last viewed by client')) {
                        last_viewed_by_client = value;
                    } else if (title.includes('hires')) {
                        hires = value;
                    } else if (title.includes('interviewing')) {
                        interviewing = value;
                    } else if (title.includes('invites sent')) {
                        invites_sent = value;
                    } else if (title.includes('unanswered invites')) {
                        unanswered_invites = value;
                    }
                }
            });
        }
    } catch (e) { }

    // Get skills - try to find and expand skills section
    let skills = [];

    try {
        const skillsSection = document.evaluate(
            '//*[@id="main"]/div/div/div[1]/div/div/section[3]',
            document,
            null,
            XPathResult.FIRST_ORDERED_NODE_TYPE,
            null
        ).singleNodeValue;

        if (skillsSection) {
            // Get all visible skills
            const allSkillElements = skillsSection.querySelectorAll('[data-test="Skill"] .air3-line-clamp');
            skills = Array.from(allSkillElements)
                .map(el => el.textContent.trim())
                .filter(text => text && !text.includes('more'));

            // Try to expand and get hidden skills
            const expandButtons = skillsSection.querySelectorAll('.air3-btn-secondary[role="button"]');
            expandButtons.forEach(button => {
                if (button.textContent.includes('more')) {
                    try {
                        button.click();
                        setTimeout(() => {
                            const hiddenSkills = skillsSection.querySelectorAll('.air3-popover [data-test="Skill"] .air3-line-clamp');
                            hiddenSkills.forEach(skillElement => {
                                const skillText = skillElement.textContent.trim();
                                if (skillText && !skills.includes(skillText)) {
                                    skills.push(skillText);
                                }
                            });
                        }, 100);
                    } catch (e) { }
                }
            });

            // Remove duplicates
            skills = [...new Set(skills)];
        }
    } catch (e) { }


    // Get client section and extract detailed information
    const clientSection = document.evaluate(
        '//*[@id="main"]/div/div/div[1]/div/div/div[1]',
        document,
        null,
        XPathResult.FIRST_ORDERED_NODE_TYPE,
        null
    ).singleNodeValue;

    // Extract client info as raw text
    let client_info_raw = null;

    // Extract detailed client information from the clientSection
    let client_location = null;
    let member_since = null;
    let total_spent = null;
    let total_hires = null;
    let total_active = null;
    let total_client_hours = null;
    let client_local_time = null;
    let client_industry = null;
    let client_company_size = null;

    // Use the existing clientSection for all client data extraction
    if (clientSection) {
        try {
            // Get member since date
            const memberSinceElement = clientSection.querySelector('[data-qa="client-contract-date"]');
            if (memberSinceElement) {
                const memberText = memberSinceElement.textContent.trim();
                // Extract date from "Member since Jan 23, 2021"
                const memberMatch = memberText.match(/Member since (.+)/);
                if (memberMatch) {
                    member_since = memberMatch[1].trim();
                }
            }

            // Get client location
            const locationElement = clientSection.querySelector('[data-qa="client-location"]');
            if (locationElement) {
                const strongElement = locationElement.querySelector('strong');
                if (strongElement) {
                    client_location = strongElement.textContent.trim();
                }
            }

            // Get local time
            const timeElement = clientSection.querySelector('[data-test="LocalTime"]');
            if (timeElement) {
                client_local_time = timeElement.textContent.trim();
            }

            // Get client industry
            const industryElement = clientSection.querySelector('[data-qa="client-company-profile-industry"]');
            if (industryElement) {
                client_industry = industryElement.textContent.trim();
            }

            // Get company size
            const companySizeElement = clientSection.querySelector('[data-qa="client-company-profile-size"]');
            if (companySizeElement) {
                client_company_size = companySizeElement.textContent.trim();
            }

            // Get total spent
            const spendElement = clientSection.querySelector('[data-qa="client-spend"]');
            if (spendElement) {
                const spendText = spendElement.textContent.trim();
                // Extract amount from "$46K total spent"
                const spendMatch = spendText.match(/(\$[\d,]+[KMB]?)/);
                if (spendMatch) {
                    total_spent = spendMatch[1];
                }
            }

            // Get hires information
            const hiresElement = clientSection.querySelector('[data-qa="client-hires"]');
            if (hiresElement) {
                const hiresText = hiresElement.textContent.trim();
                // Extract from "42 hires, 11 active"
                const hiresMatch = hiresText.match(/(\d+)\s+hires(?:,\s*(\d+)\s+active)?/);
                if (hiresMatch) {
                    total_hires = hiresMatch[1];
                    if (hiresMatch[2]) {
                        total_active = hiresMatch[2];
                    }
                }
            }

            // Get total hours
            const hoursElement = clientSection.querySelector('[data-qa="client-hours"]');
            if (hoursElement) {
                const hoursText = hoursElement.textContent.trim();
                // Extract from "811 hours"
                const hoursMatch = hoursText.match(/(\d+(?:,\d+)*)\s+hours?/);
                if (hoursMatch) {
                    total_client_hours = hoursMatch[1];
                }
            }
        } catch (e) { }
    }

    return {
        job_id: job_id,
        title: title,
        budget: budget,
        location_type: location_type,
        description: description,
        hourly_rate: hourly_rate,
        posted_date: posted_date,
        proposals_count: proposals_count,
        last_viewed_by_client: last_viewed_by_client,
        hires: hires,
        interviewing: interviewing,
        invites_sent: invites_sent,
        unanswered_invites: unanswered_invites,
        experience_level: experience_level,
        job_type: job_type,
        duration: duration,
        project_type: project_type,
        work_hours: work_hours,
        skills: skills,
        client_info_raw: client_info_raw,
        client_location: client_location,
        client_local_time: client_local_time,
        client_industry: client_industry,
        client_company_size: client_company_size,
        member_since: member_since,
        total_spent: total_spent,
        total_hires: total_hires,
        total_active: total_active,
        total_client_hours: total_client_hours
    };
}
