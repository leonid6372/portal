package news

import (
	"database/sql"
	"fmt"
	"portal/internal/storage/postgres"
	"strconv"
	"time"
)

const (
	// COALESCE() устанавливает значение update_date равное creation_date, если первое равно null, т.к. нельзя считать null в *time.Time
	qrGetPostsPage            = `SELECT post_id, title, "text", creation_date, COALESCE(update_date, creation_date) AS update_date FROM post WHERE $1 < creation_date AND creation_date < $2 LIMIT $3 OFFSET $4;`
	qrGetPostsIDByDateFilter  = `SELECT post_id FROM post WHERE $1 < creation_date AND creation_date < $2`
	qrGetPostByID             = `SELECT title, "text", creation_date, COALESCE(update_date, creation_date) AS update_date FROM post WHERE post_id = $1;`
	qrGetPostsAmount          = `SELECT count(post_id) FROM post;`
	qrGetPostText             = `SELECT "text" FROM post WHERE post_id = $1;`
	qrGetCommentsByPostID     = `SELECT comment_id, user_id, post_id, text, creation_date, COALESCE(update_date, creation_date) AS update_date FROM comment WHERE post_id = $1;`
	qrGetUncheckedComments    = `SELECT comment_id, user_id, post_id, text, creation_date, COALESCE(update_date, creation_date) AS update_date, is_checked FROM comment WHERE is_checked = FALSE;`
	qrUpdatePost              = `UPDATE post SET title = $1, "text" = $2, update_date = CURRENT_TIMESTAMP WHERE post_id = $3;`
	qrUpdateCommentText       = `UPDATE comment SET "text" = $1, update_date = CURRENT_TIMESTAMP, is_checked = FALSE WHERE comment_id = $2;`
	qrUpdateCommentIsChecked  = `UPDATE comment SET is_checked = TRUE WHERE comment_id = $1;`
	qrGetLikesAmountByPostID  = `SELECT likes_amount FROM likes_amount WHERE post_id = $1;`
	qrGetImagePathsByPostID   = `SELECT "path" FROM post_image WHERE post_id = $1;`
	qrGetTagsByPostID         = `SELECT tag_id, "name", color FROM post_tags WHERE post_id = $1;`
	qrGetTags                 = `SELECT * FROM tag;`
	qrNewTag                  = `INSERT INTO tag("name", color) VALUES ($1, $2);`
	qrNewLike                 = `INSERT INTO "like"(user_id, post_id) VALUES ($1, $2);`
	qrNewComment              = `INSERT INTO "comment"(user_id, post_id, "text", creation_date, is_checked) VALUES ($1, $2, $3, CURRENT_TIMESTAMP, FALSE);`
	qrNewPost                 = `INSERT INTO post(title, "text", creattion_date, update_date) VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING post_id;`
	qrNewPostImage            = `INSERT INTO post_image(post_id, "path") VALUES ($1, $2);`
	qrNewInPostTag            = `INSERT INTO in_post_tag(post_id, tag_id) VALUES ($1, $2);`
	qrDeleteInPostTagByPostID = `DELETE FROM in_post_image WHERE post_id = $1;`
	qrDeleteComment           = `DELETE FROM comment WHERE comment_id = $1;`
	qrDeletePost              = `DELETE FROM post WHERE post_id = $1;`
	qrDeletePostImageByPostID = `DELETE FROM post_image WHERE post_id = $1;`
)

const (
	postsPerPage = 15 // количество записей на странице
)

type Post struct {
	PostID       int       `json:"post_id,omitempty"`
	Title        string    `json:"title,omitempty"`
	Text         string    `json:"text,omitempty"`
	CreationDate time.Time `json:"creation_date,omitempty"`
	UpdateDate   time.Time `json:"update_date,omitempty"`
}

// Also set created post id value to p.PostID
func (p *Post) NewPost(storage *postgres.Storage, title, text string) error {
	const op = "storage.postgres.entities.news.NewPost"

	err := storage.DB.QueryRow(qrNewPost, title, text).Scan(&p.PostID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Post) GetText(storage *postgres.Storage, postID int) error {
	const op = "storage.postgres.entities.news.GetText"

	err := storage.DB.QueryRow(qrGetPostText, postID).Scan(&p.Text)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Post) UpdatePost(storage *postgres.Storage, title, text string, postID int) error {
	const op = "storage.postgres.entities.news.UpdatePost"

	_, err := storage.DB.Exec(qrUpdatePost, title, text, postID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Post) DeletePost(storage *postgres.Storage, postID int) error {
	const op = "storage.postgres.entities.news.DeletePost"

	_, err := storage.DB.Exec(qrDeletePost, postID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

type PostsPage struct {
	Posts      []Post     `json:"posts,omitempty"`
	Pagination Pagination `json:"pagination,omitempty"`
}

// Return slice of Article structs with empty values of Images and Tags
func (p *Post) GetPostsPage(storage *postgres.Storage, tags []string, page int, createdAfter, createdBefore time.Time) ([]Post, error) {
	const op = "storage.postgres.entities.news.GetPostsPage"

	var ps []Post

	var qrResult *sql.Rows
	var err error
	var postsAmount int

	// Make pagination
	if page < 0 {
		return nil, fmt.Errorf("%s: page in out of range", op)
	}
	if page == 0 {
		page = 1
	}
	limit := postsPerPage
	offset := limit * (page - 1)

	// If there are no tags, get posts without filter
	// Else get posts with filter
	if len(tags) == 0 {
		qrResult, err = storage.DB.Query(qrGetPostsPage, createdAfter, createdBefore, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		for qrResult.Next() {
			if err := qrResult.Scan(&p.PostID, &p.Title, &p.Text, &p.CreationDate, &p.UpdateDate); err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			ps = append(ps, *p)
		}

		// Count posts amount to make MaxPage in pagination
		if err := storage.DB.QueryRow(qrGetPostsAmount).Scan(&postsAmount); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	} else {
		// Build SQL query with filter
		qrGetPostIDsByTags := `SELECT post_id FROM post_tags WHERE post_id IN(` + qrGetPostsIDByDateFilter + `) AND `
		for i, tag := range tags {
			if i > 0 {
				qrGetPostIDsByTags += ` OR `
			}
			qrGetPostIDsByTags += ` "name" = '` + tag + `' `
		}
		qrGetPostIDsByTags += ` GROUP BY post_id HAVING COUNT(post_id) >= ` + strconv.Itoa(len(tags))

		// Count posts amount to make MaxPage in pagination
		qrGetPostsWithTagsAmount := `SELECT COUNT(post_id) FROM ( `
		qrGetPostsWithTagsAmount += qrGetPostIDsByTags + `) AS TEMP_TABLE;`
		if err := storage.DB.QueryRow(qrGetPostsWithTagsAmount, createdAfter, createdBefore).Scan(&postsAmount); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		qrGetPostIDsByTags += ` LIMIT ` + strconv.Itoa(limit) + ` OFFSET ` + strconv.Itoa(offset) + `;`

		// Get all post ID with filter
		qrResult, err = storage.DB.Query(qrGetPostIDsByTags, createdAfter, createdBefore)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		// Fills Post struct for each post ID
		for qrResult.Next() {
			if err := qrResult.Scan(&p.PostID); err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			if err := storage.DB.QueryRow(qrGetPostByID, p.PostID).Scan(&p.Title, &p.Text, &p.CreationDate, &p.UpdateDate); err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			ps = append(ps, *p)
		}
	}

	// Pagination is needless if there is no posts by filter
	if postsAmount == 0 {
		return ps, nil
	}
	// Else make Pagination struct
	var pagination Pagination
	if err := pagination.NewPagination(postsAmount, limit, page); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return ps, nil
}

type PostImage struct {
	PostImageID int    `json:"post_image_id,omitempty"`
	PostID      int    `json:"post_id,omitempty"`
	Path        string `json:"path,omitempty"`
}

func (pi *PostImage) NewPostImage(storage *postgres.Storage, postID int, path string) error {
	const op = "storage.postgres.entities.news.NewPostImage"

	_, err := storage.DB.Exec(qrNewPostImage, postID, path)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (pi *PostImage) GetImagePathsByPostID(storage *postgres.Storage, postID int) ([]string, error) {
	const op = "storage.postgres.entities.news.GetImagePathsByPostID"

	qrResult, err := storage.DB.Query(qrGetImagePathsByPostID, postID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var paths []string
	for qrResult.Next() {
		var path string
		if err := qrResult.Scan(&path); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		paths = append(paths, path)
	}

	return paths, nil
}

func (pi *PostImage) DeletePostImageByPostID(storage *postgres.Storage, postID int) error {
	const op = "storage.postgres.entities.news.DeletePostImageByPostID"

	_, err := storage.DB.Exec(qrDeletePostImageByPostID, postID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

type Tag struct {
	TagID int    `json:"tag_id,omitempty"`
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

func (t *Tag) NewTag(storage *postgres.Storage, name, color string) error {
	const op = "storage.postgres.entities.news.NewTag"

	_, err := storage.DB.Exec(qrNewTag, name, color)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (t *Tag) GetTags(storage *postgres.Storage) ([]Tag, error) {
	const op = "storage.postgres.entities.news.GetTags"

	qrResult, err := storage.DB.Query(qrGetTags)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tags := []Tag{}
	for qrResult.Next() {
		if err := qrResult.Scan(&t.TagID, &t.Name, &t.Color); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		tags = append(tags, *t)
	}

	return tags, nil
}

func (t *Tag) GetTagsByPostID(storage *postgres.Storage, postID int) ([]Tag, error) {
	const op = "storage.postgres.entities.news.GetTagsByPostID"

	qrResult, err := storage.DB.Query(qrGetTagsByPostID, postID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var tags []Tag
	for qrResult.Next() {
		if err := qrResult.Scan(&t.TagID, &t.Name, &t.Color); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		tags = append(tags, *t)
	}

	return tags, nil
}

type InPostTag struct {
	InPostTagID int `json:"in_post_tag_id,omitempty"`
	TagID       int `json:"tag_id,omitempty"`
	PostID      int `json:"post_id,omitempty"`
}

func (ipt *InPostTag) NewInPostTag(storage *postgres.Storage, postID, tagID int) error {
	const op = "storage.postgres.entities.news.NewInPostTag"

	_, err := storage.DB.Exec(qrNewInPostTag, postID, tagID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (ipt *InPostTag) DeleteInPostTagByPostID(storage *postgres.Storage, postID int) error {
	const op = "storage.postgres.entities.news.DeleteInPostTagByPostID"

	_, err := storage.DB.Exec(qrDeleteInPostTagByPostID, postID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

type Like struct {
	UserID int `json:"user_id,omitempty"`
	PostID int `json:"post_id,omitempty"`
}

func (l *Like) NewLike(storage *postgres.Storage, userID, postID int) error {
	const op = "storage.postgres.entities.news.NewLike"

	_, err := storage.DB.Exec(qrNewLike, userID, postID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (l *Like) GetLikesAmount(storage *postgres.Storage, postID int) (int, error) {
	const op = "storage.postgres.entities.news.GetLikesAmount"

	var amount int

	// Запрашиваем кол-во лайков у поста в БД. Если лайки ещё не ставили или post_id нет, то вернётся 0
	err := storage.DB.QueryRow(qrGetLikesAmountByPostID, postID).Scan(&amount)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return amount, nil
}

type Comment struct {
	CommentID    int       `json:"comment_id,omitempty"`
	UserID       int       `json:"user_id,omitempty"`
	PostID       int       `json:"post_id,omitempty"`
	Text         string    `json:"text,omitempty"`
	CreationDate time.Time `json:"creation_date,omitempty"`
	UpdateDate   time.Time `json:"update_date,omitempty"`
	IsChecked    bool      `json:"is_checked,omitempty"`
}

func (c *Comment) NewComment(storage *postgres.Storage, text string, userID, postID int) error {
	const op = "storage.postgres.entities.news.NewComment"

	_, err := storage.DB.Exec(qrNewComment, userID, postID, text)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Comment) GetCommentsByPostID(storage *postgres.Storage, postID int) ([]Comment, error) {
	const op = "storage.postgres.entities.news.GetCommentsByPostID"

	qrResult, err := storage.DB.Query(qrGetCommentsByPostID, postID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var cs []Comment
	for qrResult.Next() {
		if err := qrResult.Scan(&c.CommentID, &c.UserID, &c.PostID, &c.Text, &c.CreationDate, &c.UpdateDate, &c.IsChecked); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		cs = append(cs, *c)
	}

	return cs, nil
}

func (c *Comment) GetUncheckedComments(storage *postgres.Storage) ([]Comment, error) {
	const op = "storage.postgres.entities.news.GetUncheckedComments"

	qrResult, err := storage.DB.Query(qrGetUncheckedComments)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var cs []Comment
	for qrResult.Next() {
		if err := qrResult.Scan(&c.CommentID, &c.UserID, &c.PostID, &c.Text, &c.CreationDate, &c.UpdateDate, &c.IsChecked); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		cs = append(cs, *c)
	}

	return cs, nil
}

func (c *Comment) UpdateCommentText(storage *postgres.Storage, commentID int, text string) error {
	const op = "storage.postgres.entities.news.UpdateCommentText"

	_, err := storage.DB.Exec(qrUpdateCommentText, text, commentID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Comment) UpdateCommentIsChecked(storage *postgres.Storage, commentID int) error {
	const op = "storage.postgres.entities.news.UpdateCommentIsChecked"

	_, err := storage.DB.Exec(qrUpdateCommentIsChecked, commentID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Comment) DeleteComment(storage *postgres.Storage, commentID int) error {
	const op = "storage.postgres.entities.news.DeleteComment"

	_, err := storage.DB.Exec(qrDeleteComment, commentID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

type Pagination struct {
	Next          int `json:"next"`
	Previous      int `json:"previous"`
	RecordPerPage int `json:"record_per_page"`
	CurrentPage   int `json:"current_page"`
	TotalPage     int `json:"total_page"`
}

// Generated Pagination Meta data
func (p *Pagination) NewPagination(recordsCount, limit, page int) error {
	const op = "storage.postgres.entities.news.NewPagination"

	total := (recordsCount / limit)

	// Calculator Total Page
	remainder := (recordsCount % limit)
	if remainder == 0 {
		p.TotalPage = total
	} else {
		p.TotalPage = total + 1
	}

	if page > p.TotalPage {
		return fmt.Errorf("%s: page in out of range", op)
	}

	// Set current/record per page meta data
	p.CurrentPage = page
	p.RecordPerPage = limit

	// Calculator the Next/Previous Page
	if page <= 0 {
		p.Next = page + 1
	} else if page < p.TotalPage {
		p.Previous = page - 1
		p.Next = page + 1
	} else if page == p.TotalPage {
		p.Previous = page - 1
		p.Next = 0
	}

	return nil
}
