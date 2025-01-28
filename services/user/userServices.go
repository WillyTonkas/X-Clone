package user

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"main/constants"
	"main/models"
	"regexp"
)

func CreateAccount(db *gorm.DB, username, password, mail string, location *string) error {
	if password == constants.EMPTY || username == constants.EMPTY {
		return errors.New("fields must not be empty")
	}

	var currentUser = models.User{
		Model:    gorm.Model{},
		Username: username,
		Mail:     mail,
		Location: location,
		Password: password,
	}

	db.Model(models.User{}).Create(&currentUser)

	return nil
}

func FollowAccount(db *gorm.DB, followingUserID, followedUserID uint) error {
	if followingUserID == followedUserID {
		return errors.New("invalid Id")
	}
	if alreadyFollows(db, followingUserID, followedUserID) {
		return errors.New("user already follows")
	}

	db.Model(models.Follow{}).Create(models.Follow{
		Model:           gorm.Model{},
		FollowingUserID: followingUserID,
		FollowedUserID:  followedUserID,
	})

	return nil
}

func UnfollowAccount(db *gorm.DB, followingUserID, followedUserID uint) error {
	if followingUserID == followedUserID {
		return errors.New("invalid Id")
	}

	var user models.Follow
	db.Model(models.Follow{}).First(&user, "FollowingUserID = ? AND FollowedUserID = ?", followingUserID, followedUserID)
	db.Model(models.Follow{}).Delete(&user)

	return nil
}

func ToggleLike(db *gorm.DB, userID uint, parentID uint) error {
	if !userExists(db, userID) {
		return errors.New(constants.ERRNOUSER)
	}

	var currentUser models.Like
	if isLiked(db, userID, parentID) {
		db.Model(models.Like{}).First(&currentUser, "UserID = ? AND ParentID = ?", userID, parentID)
		db.Model(models.Like{}).Delete(&currentUser)
	} else {
		db.Model(models.Like{}).Create(models.Like{
			Model:    gorm.Model{},
			ParentID: parentID,
			UserID:   userID,
		})
	}

	return nil
}

func SearchUserByUsername(db *gorm.DB, username string) ([]models.User, error) {
	var users []models.User
	result := db.Where("Username LIKE ?", username).First(&users)
	if result.RowsAffected == 0 {
		return nil, errors.New(constants.ERRNOUSER)
	}
	return users, nil
}

func SearchPostsByKeywords(db *gorm.DB, keyword string) ([]models.Post, error) {
	var posts []models.Post
	result := db.Where("Body ILIKE ?", "%"+keyword+"%").Find(&posts)
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf(constants.ERRNOPOST+" keyword used: %s", keyword)
	}
	return posts, nil
}

func GetAllPosts(db *gorm.DB) ([]models.Post, error) {
	var posts []models.Post
	// TODO SCALE THIS
	result := db.Find(&posts)
	if result.Error != nil {
		return nil, fmt.Errorf("internal server error: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, errors.New(constants.ERRNOPOST)
	}
	return posts, nil
}

func GetAllPostsByUserID(db *gorm.DB, userID uint) ([]models.Post, error) {
	var posts []models.Post
	result := db.Where("user_id = ?", userID).Find(&posts)
	if result.Error != nil {
		return nil, fmt.Errorf("internal server error: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, errors.New(constants.ERRNOPOST)
	}
	return posts, nil
}

func CreatePost(db *gorm.DB, userID uint, parentID *uint, quoteID *uint, body string) error {
	if !userExists(db, userID) {
		return errors.New(constants.ERRNOUSER)
	}
	post := models.Post{
		UserID:   userID,
		ParentID: parentID,
		Quote:    quoteID,
		Body:     body,
	}

	return db.Create(&post).Error
}

// AUX.

func MailAlreadyUsed(db *gorm.DB, mail string) bool {
	var user models.User
	err := db.Where("Mail = ?", mail).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	} else if err != nil {
		log.Printf("Error querying user by email: %v", err)
		return false
	}
	return true
}

func UsernameAlreadyUsed(db *gorm.DB, username string) bool {
	var user models.User
	err := db.Where("Username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	} else if err != nil {
		log.Printf("Error querying user by username: %v", err)
		return false
	}
	return true
}

func ValidateCredentials(db *gorm.DB, inputUser, password string) bool {
	var user models.User

	field := "Mail"
	if !IsEmail(inputUser) {
		field = "Username"
	}
	err := queryUserByField(db, field, inputUser, password, &user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
		log.Printf("Error querying user by %s: %v", field, err)
		return false
	}
	return true
}

func IsEmail(email string) bool {
	re := regexp.MustCompile(constants.EMAILREGEXPATTERNS)
	return re.MatchString(email)
}

func GetUserByID(db *gorm.DB, userID uint) (models.User, error) {
	var user models.User
	err := db.First(&user, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user, errors.New(constants.ERRNOUSER)
		}
		return user, errors.New("failed to retrieve the user from the database")
	}
	return user, nil
}

func GetPostByID(db *gorm.DB, postID uint) (models.Post, error) {
	var post models.Post
	err := db.First(&post, postID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return post, errors.New(constants.ERRNOPOST)
		}
		return post, errors.New("failed to retrieve the post from the database")
	}
	return post, nil
}

func UpdateProfile(db *gorm.DB, user *models.User) error {
	return db.Save(user).Error
}

func GetFollowers(db *gorm.DB, userID uint) ([]models.User, error) {
	var followers []models.User
	result := db.Table("users").
		Select("users.*").
		Joins("JOIN follows ON users.id = follows.following_user_id").
		Where("followed_user_id = ?", userID).
		Find(&followers)

	if result.RowsAffected == 0 {
		return nil, errors.New(constants.ERRNOUSER)
	}
	return followers, nil
}

func GetFollowing(db *gorm.DB, u uint) ([]models.User, error) {
	var following []models.User
	result := db.Table("users").
		Select("users.*").
		Joins("JOIN follows ON users.id = follows.followed_user_id").
		Where("following_user_id = ?", u).
		Find(&following)

	if result.RowsAffected == 0 {
		return nil, errors.New(constants.ERRNOUSER)
	}
	return following, nil
}

func queryUserByField(db *gorm.DB, field, value, password string, user *models.User) error {
	return db.Where(fmt.Sprintf("%s = ? AND Password = ?", field), value, password).First(user).Error
}

func alreadyFollows(db *gorm.DB, followingUserID, followedUserID uint) bool {
	return db.Model(models.Follow{}).
		Where(models.Follow{}, "FollowingUserID = ? AND FollowedUserID = ?", followingUserID, followedUserID).Error == nil
}

func isLiked(db *gorm.DB, userID, parentID uint) bool {
	return db.Model(models.Like{}).Where("UserID = ? AND ParentID = ?", userID, parentID).Error == nil
}

func userExists(db *gorm.DB, userID uint) bool {
	var user models.User
	err := db.Where("id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	} else if err != nil {
		log.Printf("Error querying user by id: %v", err)
		return false
	}
	return true
}
