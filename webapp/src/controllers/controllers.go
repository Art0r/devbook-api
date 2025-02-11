package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"webapp/src/config"
	"webapp/src/cookies"
	"webapp/src/models"
	"webapp/src/requests"
	"webapp/src/responses"
	"webapp/src/utils"

	"github.com/gorilla/mux"
)

func LoadLoginScreen(rw http.ResponseWriter, r *http.Request) {
	cookie, _ := cookies.Read(r)
	if cookie["token"] != "" {
		http.Redirect(rw, r, "/home", http.StatusFound)
		return
	}
	utils.ExecuteTemplate(rw, "login.html", nil)
}

func LoadUserSigninPage(rw http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(rw, "signup.html", nil)
}

func LoadUserPage(rw http.ResponseWriter, r *http.Request) {
	nameOrNick := strings.ToLower(r.URL.Query().Get("user"))
	url := fmt.Sprintf("%s/user?user=%s", config.ApiUrl, nameOrNick)

	response, err := requests.MakeAuthRequest(r, http.MethodGet, url, nil)
	if err != nil {
		responses.JSON(rw, http.StatusInternalServerError, responses.Error{Error: err.Error()})
		return
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		responses.CatchErrorStatusCode(rw, response)
		return
	}

	var users []models.User
	if err = json.NewDecoder(response.Body).Decode(&users); err != nil {
		responses.JSON(rw, http.StatusUnprocessableEntity, responses.Error{Error: err.Error()})
		return
	}

	utils.ExecuteTemplate(rw, "users.html", users)
}

func LoadUserProfile(rw http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	uid, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		responses.JSON(rw, http.StatusBadRequest, responses.Error{Error: err.Error()})
		return
	}

	cookie, _ := cookies.Read(r)
	loggedUid, _ := strconv.ParseUint(cookie["id"], 10, 64)

	if uid == loggedUid {
		http.Redirect(rw, r, "/profile", http.StatusFound)
		return
	}

	user, err := models.SearchForUser(uid, r)
	if err != nil {
		responses.JSON(rw, http.StatusInternalServerError, responses.Error{Error: err.Error()})
		return
	}

	utils.ExecuteTemplate(rw, "user.html", struct {
		User         models.User
		LoggedUserId uint64
	}{
		User:         user,
		LoggedUserId: loggedUid,
	})
}

func LoadLoggedUserProfile(rw http.ResponseWriter, r *http.Request) {
	cookie, _ := cookies.Read(r)
	uid, _ := strconv.ParseUint(cookie["id"], 10, 64)

	user, err := models.SearchForUser(uid, r)
	if err != nil {
		fmt.Println(err)
		responses.JSON(rw, http.StatusInternalServerError, responses.Error{Error: err.Error()})
		return
	}

	utils.ExecuteTemplate(rw, "profile.html", user)
}

func LoadLoggedUserProfileEdit(rw http.ResponseWriter, r *http.Request) {
	cookie, _ := cookies.Read(r)
	uid, _ := strconv.ParseUint(cookie["id"], 10, 64)

	ch := make(chan models.User)
	go models.SearchUserData(ch, uid, r)
	user := <-ch

	if user.Id == 0 {
		responses.JSON(rw, http.StatusInternalServerError, responses.Error{Error: "Erro ao buscar usuário"})
		return
	}

	utils.ExecuteTemplate(rw, "edit-user.html", user)
}

func LoadUpdatePasswordPage(rw http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(rw, "update-password.html", nil)
}
