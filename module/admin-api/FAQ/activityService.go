package FAQ

import (
	"errors"

	"emotibot.com/emotigo/module/admin-api/ApiError"
)

func GetQuestionLabels(appid string) ([]*Label, error) {
	labels, err := getQuestionLabels(appid)
	if err != nil {
		return nil, err
	}
	countMap, err := getAllLabelActivityCount(appid)
	if err != nil {
		return nil, err
	}

	for _, label := range labels {
		if cnt, ok := countMap[label.ID]; ok {
			label.ActivityCount = cnt
		}
	}

	return labels, nil
}

func AddNewLabel(appid string, newLabel *Label) (int, error) {
	_, err := getQuestionLabelByName(appid, newLabel.Name)
	if err == nil {
		return ApiError.REQUEST_ERROR, errors.New("label existed")
	}
	err = addQuestionLabel(appid, newLabel)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func UpdateLabel(appid string, newLabel *Label) (int, error) {
	_, err := getQuestionLabelByID(appid, newLabel.ID)
	if err != nil {
		return ApiError.REQUEST_ERROR, errors.New("label not existed")
	}
	_, err = getQuestionLabelByName(appid, newLabel.Name)
	if err == nil {
		return ApiError.REQUEST_ERROR, errors.New("new name existed")
	}
	err = updateQuestionLabel(appid, newLabel)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func DeleteLabel(appid string, id int) (int, error) {
	_, err := getQuestionLabelByID(appid, id)
	if err != nil {
		return ApiError.REQUEST_ERROR, errors.New("label not existed")
	}
	count, err := getLabelActivityCount(appid, id)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	if count > 0 {
		return ApiError.REQUEST_ERROR, errors.New("Label link to activities")
	}
	err = deleteQuestionLabel(appid, id)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func GetActivities(appid string) ([]*Activity, error) {
	activities, err := getActivities(appid)
	if err != nil {
		return nil, err
	}

	activityLabel, err := getActivityLabel(appid)
	if err != nil {
		return nil, err
	}

	for _, activity := range activities {
		if label, ok := activityLabel[activity.ID]; ok {
			labelID := label
			activity.LinkLabel = &labelID
		}
	}
	return activities, nil
}

func AddActivity(appid string, activity *Activity) (int, error) {
	// check label existed first
	if activity.LinkLabel != nil {
		label := *activity.LinkLabel
		_, err := getQuestionLabelByID(appid, label)
		if err != nil {
			return ApiError.REQUEST_ERROR, errors.New("Label not existed")
		}
	}

	// add activity first to get new activity ID
	err := addActivity(appid, activity)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func UpdateActivity(appid string, activity *Activity) (int, error) {
	// check label existed first
	if activity.LinkLabel != nil {
		label := *activity.LinkLabel
		_, err := getQuestionLabelByID(appid, label)
		if err != nil {
			return ApiError.REQUEST_ERROR, errors.New("Label not existed")
		}
	}

	// update activity
	err := updateActivity(appid, activity)
	if err != nil {
		return ApiError.DB_ERROR, err
	}

	return ApiError.SUCCESS, nil
}

func UpdateActivityStatus(appid string, id int, status bool) (int, error) {
	err := setActivityStatus(appid, id, status)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func DeleteActivity(appid string, id int) (int, error) {
	_, err := getActivityByID(appid, id)
	if err != nil {
		return ApiError.REQUEST_ERROR, errors.New("activity doesn't existed")
	}
	err = deleteActivity(appid, id)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	err = unlinkActivityLabel(appid, id)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func GetActivityByLabelID(appid string, lid int) (int, *Activity, error) {
	// check label existed first
	_, err := getQuestionLabelByID(appid, lid)
	if err != nil {
		return ApiError.REQUEST_ERROR, nil, errors.New("Label not existed")
	}

	aid, err := getActivityOfLabel(appid, lid)
	if err != nil {
		return ApiError.DB_ERROR, nil, err
	}
	activity, err := getActivityByID(appid, aid)
	if err != nil {
		return ApiError.DB_ERROR, nil, err
	}
	return ApiError.SUCCESS, activity, nil
}
