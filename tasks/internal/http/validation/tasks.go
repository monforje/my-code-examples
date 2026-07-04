package validation

import (
	"github.com/google/uuid"
)

var (
	// ValidTaskTypes - допустимые типы задач
	ValidTaskTypes = map[string]bool{
		"backend":  true,
		"frontend": true,
	}

	// ValidLevels - допустимые уровни
	ValidLevels = map[string]bool{
		"junior": true,
		"middle": true,
		"senior": true,
	}
)

// ValidateCreateTaskRequest - валидация запроса на создание задачи
func ValidateCreateTaskRequest(title, specificationMDText, description, taskType, level string, tagIDs, languageIDs []string) (string, string, string, string, string, []uuid.UUID, []uuid.UUID, error) {
	normalizedTitle, err := Title(title).Validate()
	if err != nil {
		return "", "", "", "", "", nil, nil, err
	}

	if !ValidTaskTypes[taskType] {
		return "", "", "", "", "", nil, nil, ErrInvalidTaskType
	}

	if !ValidLevels[level] {
		return "", "", "", "", "", nil, nil, ErrInvalidLevel
	}

	parsedTags, err := parseUUIDs(tagIDs)
	if err != nil {
		return "", "", "", "", "", nil, nil, err
	}

	parsedLangs, err := parseUUIDs(languageIDs)
	if err != nil {
		return "", "", "", "", "", nil, nil, err
	}

	return string(normalizedTitle), specificationMDText, description, taskType, level, parsedTags, parsedLangs, nil
}

// ValidateUpdateTaskRequest - валидация запроса на обновление задачи
func ValidateUpdateTaskRequest(title, specificationMDText, description, taskType, level string, tagIDs, languageIDs []string) (*string, *string, *string, *string, *string, *[]uuid.UUID, *[]uuid.UUID, error) {
	var normalizedTitle *string
	if title != "" {
		t, err := Title(title).Validate()
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		s := string(t)
		normalizedTitle = &s
	}

	var normalizedSpec *string
	if specificationMDText != "" {
		normalizedSpec = &specificationMDText
	}

	var normalizedDesc *string
	if description != "" {
		normalizedDesc = &description
	}

	var normalizedTaskType *string
	if taskType != "" {
		if !ValidTaskTypes[taskType] {
			return nil, nil, nil, nil, nil, nil, nil, ErrInvalidTaskType
		}
		normalizedTaskType = &taskType
	}

	var normalizedLevel *string
	if level != "" {
		if !ValidLevels[level] {
			return nil, nil, nil, nil, nil, nil, nil, ErrInvalidLevel
		}
		normalizedLevel = &level
	}

	var parsedTags *[]uuid.UUID
	if tagIDs != nil {
		tags, err := parseUUIDs(tagIDs)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		parsedTags = &tags
	}

	var parsedLangs *[]uuid.UUID
	if languageIDs != nil {
		langs, err := parseUUIDs(languageIDs)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		parsedLangs = &langs
	}

	return normalizedTitle, normalizedSpec, normalizedDesc, normalizedTaskType, normalizedLevel, parsedTags, parsedLangs, nil
}

// ValidateTaskID - валидация UUID задачи
func ValidateTaskID(id string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, ErrTaskIDInvalid
	}
	return parsed, nil
}

func parseUUIDs(ids []string) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, ErrInvalidUUID
		}
		result = append(result, parsed)
	}
	return result, nil
}
