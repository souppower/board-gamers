package board_gamers

import (
	"bytes"
	"encoding/json"
	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dghubble/sessions"
	"github.com/garyburd/go-oauth/oauth"
	"io/ioutil"
)

const (
	layout         = "January 02, 2006 at 15:04PM"
	sessionName    = "board-gamers"
	sessionSecret  = "board-gamers-secret"
	sessionUserKey = "twitterID"
)

var sessionStore = sessions.NewCookieStore([]byte(sessionSecret), nil)

var oauthClient = oauth.Client{
	TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
	ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authorize",
	TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
}

type Tweet struct {
	UserName    string
	Text        string
	LinkToTweet string
	CreatedAt   string
}

type Values struct {
	Value1 string `json:"value1"`
	Value2 string `json:"value2"`
}

type ArrivalOfGames struct {
	Id        int64     `datastore:"-" goon:"id" json:"id"`
	Shop      string    `json:"shop"`
	Games     []string  `json:"games"`
	CreatedAt time.Time `json:"createdAt"`
	Url       string    `json:"url" datastore:",noindex"`
}

type Config struct {
	TwitterConsumerKey    string
	TwitterConsumerSecret string
}

type User struct {
	UserId          string   `json:"userId" goon:"id"`
	ScreenName      string   `json:"screenName" datastore:",noindex"`
	Shops           []string `json:"shops"`
	NotificationKey string   `json:"notificationKey"`
}

type Shop struct {
	Name             string   `json:"name"`
	NotificationKeys []string `json:"notificationKeys"`
}

func init() {
	b, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(b, &oauthClient.Credentials); err != nil {
		panic(err)
	}

	http.HandleFunc("/webhook/trickplay", trickplayHandler)

	http.HandleFunc("/api/v1/arrivalOfGames", ArrivalOfGamesHandler)
	http.HandleFunc("/api/v1/user", UserHandler)
	http.HandleFunc("/api/v1/auth", AuthHandler)

	http.HandleFunc("/twitter/login", twitterLoginHandler)
	http.HandleFunc("/twitter/callback", twitterCallbackHandler)
}

func trickplayHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	g := goon.NewGoon(r)

	decoder := json.NewDecoder(r.Body)
	var t Tweet
	err := decoder.Decode(&t)
	if err != nil {
		http.Error(w, "json parse error", 500)
		log.Errorf(ctx, "json parse error: %v", err)
	}
	//TODO 入荷した判定をする
	if !strings.Contains(t.Text, "入荷しております") {
		log.Infof(ctx, "no nyuuka")
		return
	}

	log.Infof(ctx, "this is 入荷 tweet: "+t.Text)

	//TODO 入荷した商品名を抽出する
	//TODO 全ての、の後ろにスペースを挿入する
	re := regexp.MustCompile("、?「(.+?)」|、?([^「」]+拡張「.+?」)|、?[^「」]+「(.+?)」")
	submatch := re.FindAllStringSubmatch(t.Text, -1)
	var games []string
	for _, v := range submatch {
		if v[1] != "" {
			games = append(games, v[1])
		} else if v[2] != "" {
			games = append(games, v[2])
		} else if v[3] != "" {
			games = append(games, v[3])
		}

	}
	log.Infof(ctx, "%v", games)

	createdAt, err := time.Parse(layout, t.CreatedAt)
	if err != nil {
		log.Errorf(ctx, "Time Parse error: %v", err)
		return
	}
	a := &ArrivalOfGames{
		Shop:      "トリックプレイ",
		Games:     games,
		CreatedAt: createdAt,
		Url:       t.LinkToTweet,
	}
	if _, err := g.Put(a); err != nil {
		log.Errorf(ctx, "Datastore put error: %v", err)
		return
	}

	postToIOS(ctx, a)
}

func twitterLoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	httpClient := urlfetch.Client(ctx)
	tmpCred, err := oauthClient.RequestTemporaryCredentials(httpClient, "http://"+r.Host+"/twitter/callback", nil)
	if err != nil {
		http.Error(w, "tmpCred error", http.StatusInternalServerError)
		log.Errorf(ctx, "tmpCred error: %v", err)
		return
	}

	http.Redirect(w, r, oauthClient.AuthorizationURL(tmpCred, nil), http.StatusFound)
	return
}

func twitterCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	token := r.FormValue("oauth_token")
	tmpCred := &oauth.Credentials{
		Token:  token,
		Secret: oauthClient.Credentials.Secret,
	}
	httpClient := urlfetch.Client(ctx)
	tokenCred, v, err := oauthClient.RequestToken(httpClient, tmpCred, r.FormValue("oauth_verifier"))
	if err != nil {
		http.Error(w, "request token error", http.StatusInternalServerError)
		log.Errorf(ctx, "request token error: %v", err)
		return
	}
	log.Infof(ctx, "token cred: %v", tokenCred)
	log.Infof(ctx, "url.Values: %v", v)

	// セッションに保存
	session := sessionStore.New(sessionName)
	session.Values[sessionUserKey] = v["user_id"][0]
	session.Save(w)

	// ユーザIDを保存する
	u := &User{
		UserId:     v["user_id"][0],
		ScreenName: v["screen_name"][0],
	}
	log.Infof(ctx, "user: %v", u)
	g := goon.NewGoon(r)
	if _, err = g.Put(u); err != nil {
		log.Errorf(ctx, "goon put error: %v", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
	return
}

//func indexHandler(w http.ResponseWriter, r *http.Request) {
//	ctx := appengine.NewContext(r)
//
//	session, err := sessionStore.Get(r, sessionName)
//	if err == nil {
//		id := session.Values[sessionUserKey]
//		log.Infof(ctx, "id: %v", id)
//	}
//
//	log.Infof(ctx, "Hello Index!")
//}

func isAuthenticated(req *http.Request) bool {
	if _, err := sessionStore.Get(req, sessionName); err == nil {
		return true
	}
	return false
}

func postToIOS(ctx context.Context, a *ArrivalOfGames) {
	client := urlfetch.Client(ctx)

	param := Values{
		Value1: a.Shop,
		Value2: strings.Join(a.Games, ","),
	}
	paramBytes, err := json.Marshal(param)
	if err != nil {
		log.Errorf(ctx, "json marshal error: %v", err)
		return
	}
	if err != nil {
		log.Errorf(ctx, "http request error: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = client.Do(req)
	if err != nil {
		log.Errorf(ctx, "client do error: %v", err)
		return
	}
}
