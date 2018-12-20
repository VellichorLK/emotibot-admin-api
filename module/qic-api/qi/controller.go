package qi

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/util"
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

	util.WriteJSON(w, groups)
}