package qi

import (
	"strconv"
	"net/http"
	"github.com/gorilla/mux"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

func handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	group := Group{}
	err := util.ReadJSON(r, &group)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdGroup, err := CreateGroup(&group)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, createdGroup)
}

func handleGetGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := GetGroups()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	simpleGroups := make([]SimpleGroup, len(groups), len(groups))
	for i, group := range groups {
		simpleGroup := SimpleGroup{
			ID: group.ID,
			Name: group.Name,
		}

		simpleGroups[i] = simpleGroup
	}

	util.WriteJSON(w, simpleGroups)
}

func parseID(r *http.Request) (id int64, err error) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err = strconv.ParseInt(idStr, 10, 64)
	return 
	
}

func handleGetGroup(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "id is not a number", http.StatusBadRequest)
		return
	}

	group, err := GetGroupBy(id)
	if err != nil {
		logger.Error.Printf("error while get group in handleGetGroup, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if group == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	util.WriteJSON(w, group)
	
}

func handleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	group := Group{}
	err = util.ReadJSON(r, &group)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = UpdateGroup(id, &group)
	if err != nil {
		logger.Error.Printf("error while update group in handleUpdateGroup, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = DeleteGroup(id)

	if err != nil {
		logger.Error.Printf("error while delete group in handleDeleteGroup, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}