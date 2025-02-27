package internal

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

var (
	db            *sql.DB
	err           error
	errLogin      = false
	errRegister   = false
	authenticated = false
)

func checkInternalServerError(err error, w http.ResponseWriter) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func checkAllowedMethods(methods []string, w http.ResponseWriter, r *http.Request) {
	allowed := false
	for _, m := range methods {
		if r.Method == m {
			allowed = true
		}
	}

	if !allowed {
		errMethod := errors.New("Method is not Allowed")
		http.Error(w, errMethod.Error(), http.StatusMethodNotAllowed)
	}
}

func pageNotFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
	// http.ServeFile(w, r, "templates/error.html")
}

func checkCookie(w http.ResponseWriter, r *http.Request) (bool, User) {
	c, err := r.Cookie("session_token")

	var user User

	if err != nil {
		//User is not logged in
		return false, user
	}

	sessionToken := c.Value

	err = db.QueryRow("SELECT id, email, username, avatar FROM users WHERE session=?",
		sessionToken).Scan(&user.ID, &user.Email, &user.Username, &user.Avatar)

	// checkInternalServerError(err, w)
	if err != nil {
		return false, user
	}

	return true, user
}
func createCookie(w http.ResponseWriter, email string) {
    sessionToken := uuid.NewV4().String() // Just one return value

    http.SetCookie(w, &http.Cookie{
        Name:    "session_token",
        Value:   sessionToken,
        Expires: time.Now().Add(2 * 60 * 60 * time.Second),
    })

    addSession, err := db.Prepare(`UPDATE users SET session=? WHERE email=?`)
    checkInternalServerError(err, w)
    res, err := addSession.Exec(sessionToken, email)
    checkInternalServerError(err, w)
    _, err = res.RowsAffected()
    checkInternalServerError(err, w)
}


func deleteCookie(w http.ResponseWriter, userID int) {

	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		MaxAge: -1,
	})

	deleteSession, err := db.Prepare(`
		UPDATE users SET session=?
		WHERE id=?
	`)
	checkInternalServerError(err, w)
	res, err := deleteSession.Exec("logout", userID)
	checkInternalServerError(err, w)
	_, err = res.RowsAffected()
	checkInternalServerError(err, w)
}

func formatPosts(w http.ResponseWriter, posts []Post) []Post {

	for i, post := range posts {
		err = db.QueryRow("SELECT username FROM users WHERE id=?",
			post.Author).Scan(&posts[i].AuthorName)
		checkInternalServerError(err, w)
		tempTimeArray := strings.Split(post.Timestamp, "T")
		posts[i].Timestamp = tempTimeArray[0]

		tempContentArray := strings.Split(post.Content, " ")
		tempContentString := ""
		if len(tempContentArray) > 20 {
			tempContentArray = tempContentArray[:20]
		}
		for i, str := range tempContentArray {
			if i != 0 {
				tempContentString += " "
			}
			tempContentString += str
		}
		posts[i].Content = tempContentString

		var category Category
		var categories []Category

		categoriesOfPost, err := db.Query("SELECT category_id FROM postcategories WHERE post_id=?", post.ID)
		for categoriesOfPost.Next() {
			err = categoriesOfPost.Scan(&category.ID)
			checkInternalServerError(err, w)

			err = db.QueryRow("SELECT name FROM categories WHERE id=?",
				category.ID).Scan(&category.Name)
			checkInternalServerError(err, w)

			categories = append(categories, category)

		}

		posts[i].Categories = categories

		// categoryIDs := strings.Split(post.CategoryIDs, ",")
		// for _, categoryID := range categoryIDs {
		// 	categoryName := ""
		// 	err = db.QueryRow("SELECT name FROM categories WHERE id=?",
		// 		categoryID).Scan(&categoryName)
		// 	checkInternalServerError(err, w)
		// 	posts[i].CategoryNames = append(posts[i].CategoryNames, categoryName)
		// }

	}

	return posts
}

func formatPost(w http.ResponseWriter, post Post) Post {

	err = db.QueryRow("SELECT username FROM users WHERE id=?",
		post.Author).Scan(&post.AuthorName)
	checkInternalServerError(err, w)
	tempTimeArray := strings.Split(post.Timestamp, "T")
	post.Timestamp = tempTimeArray[0]

	var category Category
	var categories []Category

	categoriesOfPost, err := db.Query("SELECT category_id FROM postcategories WHERE post_id=?", post.ID)
	for categoriesOfPost.Next() {
		err = categoriesOfPost.Scan(&category.ID)
		checkInternalServerError(err, w)

		err = db.QueryRow("SELECT name FROM categories WHERE id=?",
			category.ID).Scan(&category.Name)
		checkInternalServerError(err, w)

		categories = append(categories, category)
	}

	post.Categories = categories

	// err = db.QueryRow("SELECT name FROM categories WHERE id=?",
	// 	post.Category).Scan(&post.CategoryName)
	// checkInternalServerError(err, w)

	return post
}

func formatComments(w http.ResponseWriter, comments []Comment) []Comment {

	for i, comment := range comments {
		err = db.QueryRow("SELECT username FROM users WHERE id=?",
			comment.Author).Scan(&comments[i].AuthorName)
		checkInternalServerError(err, w)

		tempTimeArray := strings.Split(comment.Timestamp, "T")
		comments[i].Timestamp = tempTimeArray[0]

	}

	return comments
}
