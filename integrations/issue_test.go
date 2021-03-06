// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/setting"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func getIssuesSelection(htmlDoc *HtmlDoc) *goquery.Selection {
	return htmlDoc.doc.Find(".issue.list").Find("li").Find(".title")
}

func getIssue(t *testing.T, repoID int64, issueSelection *goquery.Selection) *models.Issue {
	href, exists := issueSelection.Attr("href")
	assert.True(t, exists)
	indexStr := href[strings.LastIndexByte(href, '/')+1:]
	index, err := strconv.Atoi(indexStr)
	assert.NoError(t, err, "Invalid issue href: %s", href)
	return models.AssertExistsAndLoadBean(t, &models.Issue{RepoID: repoID, Index: int64(index)}).(*models.Issue)
}

func TestNoLoginViewIssues(t *testing.T) {
	prepareTestEnv(t)

	req := NewRequest(t, "GET", "/user2/repo1/issues")
	resp := MakeRequest(req)
	assert.EqualValues(t, http.StatusOK, resp.HeaderCode)
}

func TestNoLoginViewIssuesSortByType(t *testing.T) {
	prepareTestEnv(t)

	user := models.AssertExistsAndLoadBean(t, &models.User{ID: 1}).(*models.User)
	repo := models.AssertExistsAndLoadBean(t, &models.Repository{ID: 1}).(*models.Repository)
	repo.Owner = models.AssertExistsAndLoadBean(t, &models.User{ID: repo.OwnerID}).(*models.User)

	session := loginUser(t, user.Name, "password")
	req := NewRequest(t, "GET", repo.RelLink()+"/issues?type=created_by")
	resp := session.MakeRequest(t, req)
	assert.EqualValues(t, http.StatusOK, resp.HeaderCode)

	htmlDoc, err := NewHtmlParser(resp.Body)
	assert.NoError(t, err)
	issuesSelection := getIssuesSelection(htmlDoc)
	expectedNumIssues := models.GetCount(t,
		&models.Issue{RepoID: repo.ID, PosterID: user.ID},
		models.Cond("is_closed=?", false),
		models.Cond("is_pull=?", false),
	)
	if expectedNumIssues > setting.UI.IssuePagingNum {
		expectedNumIssues = setting.UI.IssuePagingNum
	}
	assert.EqualValues(t, expectedNumIssues, issuesSelection.Length())

	issuesSelection.Each(func(_ int, selection *goquery.Selection) {
		issue := getIssue(t, repo.ID, selection)
		assert.EqualValues(t, user.ID, issue.PosterID)
	})
}

func TestNoLoginViewIssue(t *testing.T) {
	prepareTestEnv(t)

	req := NewRequest(t, "GET", "/user2/repo1/issues/1")
	resp := MakeRequest(req)
	assert.EqualValues(t, http.StatusOK, resp.HeaderCode)
}
