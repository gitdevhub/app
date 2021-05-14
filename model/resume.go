package model

import (
	"encoding/json"
	"fmt"
	"gopkg.in/guregu/null.v4"
	"math"
	"time"
)

type Resume struct {
	

func (m *JobExperienceResult) Scan(src interface{}) error {
	var err error

	if src == nil {
		return nil
	}

	switch src.(type) {
	case []byte:
		err = json.Unmarshal(src.([]byte), m)
	default:
		err = fmt.Errorf("job experience field must be a []byte, got %T instead", src)
	}
	if err != nil {
		return err
	}

	return nil
}

func GetCustomResumeByID(resumeId int64, companyId int64, language string) (*ResumeResult, error) {
	var result ResumeResult

	query := db.GetDB().Table("resumes as r")

	if language != "en" {
		query.
			Select([]string{
				"r.id",
				"r.title",
				"r.desc",
				"r.email",
				"r.phone",
				"r.linkedin",
				"r.remote_job",
				"r.min_salary",
				"currencies.name",

				"job_types.names",
				"skills.names",

				"languages.names",
				"educations.json",
				"job_experiences.json",

				"u.id",
				"u.name",
				"u.username",
				"u.image",

				"COALESCE(country_translations.name, countries.name)",
				"countries.code",

				"COALESCE(city_translations.name, cities.name)",
				"cities.geoname_id",

				"IF(resume_favorites.id > 0, 1, 0)",
			})
	} else {
		query.
			Select([]string{
				"r.id",
				"r.title",
				"r.desc",
				"r.email",
				"r.phone",
				"r.linkedin",
				"r.remote_job",
				"r.min_salary",
				"currencies.name",

				"job_types.names",
				"skills.names",

				"languages.names",
				"educations.json",
				"job_experiences.json",

				"u.id",
				"u.name",
				"u.username",
				"u.image",

				"countries.name",
				"countries.code",

				"cities.name",
				"cities.geoname_id",

				"IF(resume_favorites.id > 0, 1, 0)",
			})
	}

	query.
		Joins("INNER JOIN users u ON u.id = r.user_id").
		Joins("INNER JOIN category ON category.id = r.category_id").
		Joins("INNER JOIN category_sub ON category_sub.id = r.subcategory_id").
		Joins("LEFT JOIN countries ON countries.geoname_id = r.country_id").
		Joins("LEFT JOIN cities ON cities.geoname_id = r.city_id").
		Joins("LEFT JOIN currencies ON currencies.id = r.currency_id").
		Joins("LEFT JOIN resume_favorites ON resume_favorites.resume_id = r.id AND resume_favorites.company_id = ?", companyId)

	if language != "en" {
		query.
			Joins("LEFT JOIN category_translations ON category_translations.category_id = category.id AND category_translations.language = ?", language).
			Joins("LEFT JOIN category_sub_translations ON category_sub_translations.subcategory_id = category_sub.id AND category_sub_translations.language = ?", language).
			Joins("LEFT JOIN country_translations ON country_translations.country_id = countries.geoname_id AND country_translations.language = ?", language).
			Joins("LEFT JOIN city_translations ON city_translations.city_id = cities.geoname_id AND city_translations.language = ?", language).
			Joins(`LEFT JOIN (
				SELECT
				resume_skill.resume_id,
				GROUP_CONCAT(COALESCE(skill_translations.name, skills.name) ORDER BY COALESCE(skill_translations.name, skills.name) ASC SEPARATOR ', ') AS names
				FROM resume_skill
				INNER JOIN skills ON skills.id = resume_skill.skill_id
				LEFT JOIN skill_translations ON skill_translations.skill_id = skills.id AND skill_translations.language = ?
				GROUP BY resume_skill.resume_id
			) AS skills ON skills.resume_id = r.id`, language).
			Joins(`LEFT JOIN (
				SELECT
				resume_employment.resume_id,
				GROUP_CONCAT(COALESCE(employment_translations.name, employment.name) ORDER BY COALESCE(employment_translations.name, employment.name) ASC SEPARATOR ', ') AS names
				FROM resume_employment
				INNER JOIN employment ON employment.id = resume_employment.employment_id
				LEFT JOIN employment_translations ON employment_translations.employment_id = employment.id AND employment_translations.language = ?
				GROUP BY resume_employment.resume_id
			) AS job_types ON job_types.resume_id = r.id`, language).
			Joins(`LEFT JOIN (
				SELECT
				resume_language.resume_id,
				GROUP_CONCAT(COALESCE(language_translations.name, languages.name) ORDER BY COALESCE(language_translations, languages.name) ASC SEPARATOR ', ') AS names
				FROM resume_language
				INNER JOIN languages ON languages.id = resume_language.language_id
				LEFT JOIN language_translations ON language_translations.language_id = languages.id AND language_translations.language = ?
				GROUP BY resume_language.resume_id
			) AS job_types ON job_types.resume_id = r.id`, language)
	} else {
		query.
			Joins(`LEFT JOIN (
				SELECT
				resume_skill.resume_id,
				GROUP_CONCAT(skills.name ORDER BY skills.name ASC SEPARATOR ', ') AS names
				FROM resume_skill
				INNER JOIN skills ON skills.id = resume_skill.skill_id
				GROUP BY resume_skill.resume_id
			) AS skills ON skills.resume_id = r.id`).
			Joins(`LEFT JOIN (
				SELECT
				resume_employment.resume_id,
				GROUP_CONCAT(employment.name ORDER BY employment.name ASC SEPARATOR ', ') AS names
				FROM resume_employment
				INNER JOIN employment ON employment.id = resume_employment.employment_id
				GROUP BY resume_employment.resume_id
			) AS job_types ON job_types.resume_id = r.id`).
			Joins(`LEFT JOIN (
				SELECT
				resume_language.resume_id,
				GROUP_CONCAT(languages.name ORDER BY languages.name ASC SEPARATOR ', ') AS names
				FROM resume_language
				INNER JOIN languages ON languages.id = resume_language.language_id
				GROUP BY resume_language.resume_id
			) AS languages ON languages.resume_id = r.id`).
			Joins(`LEFT JOIN (
				SELECT
				job_experience.resume_id,
				JSON_ARRAYAGG(JSON_OBJECT(
					'title', job_experience.title,
					'desc', job_experience.desc,
					'date_from', job_experience.date_from,
					'date_to', job_experience.date_to,
					'country', (
						SELECT JSON_OBJECT(
							'code', countries.code,
							'name', countries.name
						)
						FROM countries WHERE countries.geoname_id = job_experience.country_id
					),
					'city', (
						SELECT JSON_OBJECT(
							'code', cities.geoname_id,
							'name', cities.name
						)
						FROM cities WHERE cities.geoname_id = job_experience.city_id
					),
					'employment', (
						SELECT JSON_OBJECT(
							'name', employment_job_experience.name
						)
						FROM employment_job_experience WHERE employment_job_experience.id = job_experience.employment_id
					),
					'company', (
						JSON_OBJECT(
							'name', COALESCE(
								(SELECT companies.name FROM companies WHERE companies.id = job_experience.company_id),
								job_experience.company_name
							)
						)
					)
				)) as json
				FROM job_experience
				GROUP BY job_experience.resume_id
			) AS job_experiences ON job_experiences.resume_id = r.id`).
			Joins(`LEFT JOIN (
				SELECT
				educations.resume_id,
				JSON_ARRAYAGG(JSON_OBJECT(
					'desc', educations.desc,
					'date_from', educations.date_from,
					'date_to', educations.date_to,
					'university', (
						SELECT JSON_OBJECT(
							'name', universities.name
						)
						FROM universities WHERE universities.id = educations.university_id
					),
					'department', (
						SELECT JSON_OBJECT(
							'name', department.name
						)
						FROM department WHERE department.id = educations.department_id
					),
					'degree', (
						SELECT JSON_OBJECT(
							'name', degree.name
						)
						FROM degree WHERE degree.id = educations.degree_id
					),
					'form_education', (
						SELECT JSON_OBJECT(
							'name', form_education.name
						)
						FROM form_education WHERE form_education.id = educations.form_education_id
					)
				)) as json
				FROM educations
				GROUP BY educations.resume_id
			) AS educations ON educations.resume_id = r.id`)
	}

	query.
		Where("r.id = ?", resumeId).
		Order("r.created_at desc").
		Limit(1)

	row := query.Row()

	