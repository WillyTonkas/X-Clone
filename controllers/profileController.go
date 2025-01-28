package controllers

import (
	"encoding/json"
	"gorm.io/gorm"
	"main/constants"
	"main/models"
	"main/services/user"
	"net/http"
	"strconv"
)

func ViewUserProfileHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	userID, err := strconv.Atoi(r.PathValue("userid"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, getUserErr := user.GetUserByID(db, uint(userID))
	if getUserErr != nil {
		http.Error(w, getUserErr.Error(), http.StatusBadRequest)
		return
	}

	if encodeError := json.NewEncoder(w).Encode(user); encodeError != nil {
		http.Error(w, encodeError.Error(), http.StatusInternalServerError)
		return
	}
}

func EditUserProfileHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	userID, err := strconv.Atoi(r.PathValue("userid"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var currentUser models.User
	err = json.NewDecoder(r.Body).Decode(&currentUser)
	if err != nil {
		http.Error(w, "Invalid user data", http.StatusBadRequest)
		return
	}

	if currentUser.ID != uint(userID) {
		http.Error(w, "User ID in URL does not match user ID in body", http.StatusBadRequest)
		return
	}

	if updateProfileErr := user.UpdateProfile(db, &currentUser); updateProfileErr != nil {
		http.Error(w, updateProfileErr.Error(), http.StatusBadRequest)
		return
	}

	if encodeError := json.NewEncoder(w).Encode(currentUser); encodeError != nil {
		http.Error(w, encodeError.Error(), http.StatusInternalServerError)
		return
	}
}

func ViewFollowersProfileHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	userID, err := strconv.Atoi(r.PathValue("userid"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	followers, getFollowersErr := user.GetFollowers(db, uint(userID))
	if getFollowersErr != nil {
		http.Error(w, getFollowersErr.Error(), http.StatusBadRequest)
		return
	}

	if encodeError := json.NewEncoder(w).Encode(followers); encodeError != nil {
		http.Error(w, encodeError.Error(), http.StatusInternalServerError)
		return
	}
}

func ViewFollowingProfileHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	userID, err := strconv.Atoi(r.PathValue("userid"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	following, getFollowingErr := user.GetFollowing(db, uint(userID))
	if getFollowingErr != nil {
		http.Error(w, getFollowingErr.Error(), http.StatusBadRequest)
		return
	}

	if encodeError := json.NewEncoder(w).Encode(following); encodeError != nil {
		http.Error(w, encodeError.Error(), http.StatusInternalServerError)
		return
	}
}

// ----------------------------- AUX ----------------------------- //

var ViewFollowersProfileEndpoint = models.Endpoint{
	Method:          models.GET,
	Path:            constants.BASEURL + "profile/followers/user/{userid}",
	HandlerFunction: ViewFollowersProfileHandler,
}

var ViewFollowingProfileEndpoint = models.Endpoint{
	Method:          models.GET,
	Path:            constants.BASEURL + "profile/following/user/{userid}",
	HandlerFunction: ViewFollowingProfileHandler,
}

var ViewUserProfileEndpoint = models.Endpoint{
	Method:          models.GET,
	Path:            constants.BASEURL + "profile/{userid}",
	HandlerFunction: ViewUserProfileHandler,
}

var EditUserProfileEndpoint = models.Endpoint{
	Method:          models.PUT,
	Path:            constants.BASEURL + "profile/{userid}/edit",
	HandlerFunction: EditUserProfileHandler,
}
