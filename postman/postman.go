package postman

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Config struct {
	Package           string
	TestFile          string
	PostmanFile       string
	RouterFunc        string
	Variables         map[string]string
	SetupRouter       string
	AdditionalImports string
}

func (c *Config) Generate() error {
	testData := headerText(c)
	os.Remove(c.TestFile)
	postmanData, err := os.ReadFile(c.PostmanFile)
	if err != nil {
		return err
	}
	var pm *Postman
	jsonData := string(postmanData)
	for k, e := range c.Variables {
		jsonData = strings.Replace(jsonData, fmt.Sprintf("{{%v}}", k), e, -1)
	}
	err = json.Unmarshal([]byte(jsonData), &pm)
	if err != nil {
		return err
	}

	var host string

	for _, v := range pm.Item {
		for _, item := range v.Item {
			host = item.Request.URL.Host[0]
			break
		}
		break
	}

	testData += fmt.Sprintf(`
var host string = "%v"
`, host)

	for _, v := range pm.Item {
		httpTests := []string{}
		for _, item := range v.Item {
			body := `""`
			if len(item.Request.Body.Raw) > 0 {
				body = fmt.Sprintf(`fmt.Sprint(`+"`%v`"+`)`, item.Request.Body.Raw)
			}

			caser := cases.Title(language.AmericanEnglish)
			method := caser.String(item.Request.Method)

			innerData := fmt.Sprintf(`
		{
	Name: "%v",
	URL: host + "/%v",
	Method: http.Method%v,
	Body: %v,
	ExpectedStatus: http.StatusOK,
	ExpectedContains: nil,
	}`, item.Name, strings.Join(item.Request.URL.Path, "/"), method, body)
			httpTests = append(httpTests, innerData)
		}

		testFunc := strings.Replace(v.Name, " ", "", -1)

		testData += `
func Test` + testFunc + `(t *testing.T) {
	tests := []HTTPTest{` + strings.Join(httpTests, ",") + `}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			run, err := RunHTTPTest(v)
			assert.Nil(t, err)
			body, err := io.ReadAll(run.Body)
			assert.Nil(t, err)
			assert.Equal(t, v.ExpectedStatus, run.StatusCode)
			if v.ExpectedContains != nil {
				stringBody := string(body)
				for _, c := range v.ExpectedContains {
					assert.Contains(t, stringBody, c)
				}
			}
			t.Logf("Test %v got: %v\n", v.Name, string(body))
		})
	}
}
`
	}

	fmt.Printf("Writing to test file: %v \n", c.TestFile)
	f, err := os.Create(c.TestFile)
	if err != nil {
		return err
	}
	f.WriteString(testData)

	return f.Close()
}

func headerText(c *Config) string {
	return `package ` + c.Package + `

import (
	"github.com/stretchr/testify/assert"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	` + c.AdditionalImports + `
)

// HTTPTest contains all the parameters for a HTTP Unit Test
type HTTPTest struct {
	Name string
	URL string
	Method string
	Body string
	ExpectedStatus int
	ExpectedContains []string
}

// RunHTTPTest accepts a HTTPTest type to execute the HTTP request
func RunHTTPTest(test HTTPTest) (*http.Response, error) {
	req, err := http.NewRequest(test.Method, test.URL, strings.NewReader(test.Body))
	if err != nil {
		return nil, err
	}
	rr := httptest.NewRecorder()
	` + c.SetupRouter + `
	` + c.RouterFunc + `.ServeHTTP(rr, req)
	return rr.Result(), err
}
`
}

type info struct {
	PostmanID string `json:"_postman_id"`
	Name      string `json:"name"`
	Schema    string `json:"schema"`
}

type item struct {
	Name string    `json:"name"`
	Item []subItem `json:"item"`
}

type subItem struct {
	Name     string        `json:"name"`
	Request  request       `json:"request"`
	Response []interface{} `json:"response"`
}

type request struct {
	Method string      `json:"method"`
	Header interface{} `json:"header"`
	Body   body        `json:"body"`
	URL    url         `json:"url"`
}

type url struct {
	Raw  string   `json:"raw"`
	Host []string `json:"host"`
	Path []string `json:"path"`
}

type body struct {
	Mode string `json:"mode"`
	Raw  string `json:"raw"`
}

type auth struct {
	Type   string `json:"type"`
	Bearer []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		Type  string `json:"type"`
	} `json:"bearer"`
}

type event struct {
	Listen string `json:"listen"`
	Script struct {
		ID   string   `json:"id"`
		Type string   `json:"type"`
		Exec []string `json:"exec"`
	} `json:"script"`
}

type variable struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type Postman struct {
	Info     info       `json:"info"`
	Item     []*item    `json:"item"`
	Auth     auth       `json:"auth"`
	Event    []event    `json:"event"`
	Variable []variable `json:"variable"`
}
