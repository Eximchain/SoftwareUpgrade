package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// NodeType is an enumeration, consisting of ntList and ntMap
type NodeType int

const (
	ntList NodeType = iota
	ntMap
)

type (
	// NodeData contains a string, and an integer (index) indicating where to read the next char from the given string
	NodeData struct {
		bytes string
		index int
	}
	Node struct {
		name     string
		nodetype NodeType
		data     interface{}
	}
)

type (
	ErrorResponse struct {
		Errors []struct {
			Status string `json:"status"`
			Title  string `json:"title"`
		} `json:"errors"`
	}

	TFE struct {
		organization       string
		workspaceName      string
		authorizationToken string
		id                 string
		request            *http.Request
	}

	GetWorkspaceResponse struct {
		Data struct {
			ID         string `json:"id"`
			Type       string `json:"type"`
			Attributes struct {
				Name             string    `json:"name"`
				Environment      string    `json:"environment"`
				AutoApply        bool      `json:"auto-apply"`
				Locked           bool      `json:"locked"`
				CreatedAt        time.Time `json:"created-at"`
				WorkingDirectory string    `json:"working-directory"`
				TerraformVersion string    `json:"terraform-version"`
				VcsRepo          struct {
					Branch            string `json:"branch"`
					IngressSubmodules bool   `json:"ingress-submodules"`
					Identifier        string `json:"identifier"`
					OauthTokenID      string `json:"oauth-token-id"`
					WebhookURL        string `json:"webhook-url"`
				} `json:"vcs-repo"`
				Permissions struct {
					CanUpdate         bool `json:"can-update"`
					CanDestroy        bool `json:"can-destroy"`
					CanQueueDestroy   bool `json:"can-queue-destroy"`
					CanQueueRun       bool `json:"can-queue-run"`
					CanUpdateVariable bool `json:"can-update-variable"`
					CanLock           bool `json:"can-lock"`
					CanReadSettings   bool `json:"can-read-settings"`
				} `json:"permissions"`
				Actions struct {
					IsDestroyable bool `json:"is-destroyable"`
				} `json:"actions"`
			} `json:"attributes"`
			Relationships struct {
				Organization struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
				} `json:"organization"`
				LatestRun struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
					Links struct {
						Related string `json:"related"`
					} `json:"links"`
				} `json:"latest-run"`
				CurrentRun struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
					Links struct {
						Related string `json:"related"`
					} `json:"links"`
				} `json:"current-run"`
				CurrentStateVersion struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
					Links struct {
						Related string `json:"related"`
					} `json:"links"`
				} `json:"current-state-version"`
			} `json:"relationships"`
			Links struct {
				Self string `json:"self"`
			} `json:"links"`
		} `json:"data"`
	}

	ListWorkspaceResponse struct {
		Data []struct {
			ID         string `json:"id"`
			Type       string `json:"type"`
			Attributes struct {
				Name             string    `json:"name"`
				Environment      string    `json:"environment"`
				AutoApply        bool      `json:"auto-apply"`
				Locked           bool      `json:"locked"`
				CreatedAt        time.Time `json:"created-at"`
				WorkingDirectory string    `json:"working-directory"`
				TerraformVersion string    `json:"terraform-version"`
				VcsRepo          struct {
					Branch            string `json:"branch"`
					IngressSubmodules bool   `json:"ingress-submodules"`
					Identifier        string `json:"identifier"`
					OauthTokenID      string `json:"oauth-token-id"`
					WebhookURL        string `json:"webhook-url"`
				} `json:"vcs-repo"`
				Permissions struct {
					CanUpdate         bool `json:"can-update"`
					CanDestroy        bool `json:"can-destroy"`
					CanQueueDestroy   bool `json:"can-queue-destroy"`
					CanQueueRun       bool `json:"can-queue-run"`
					CanUpdateVariable bool `json:"can-update-variable"`
					CanLock           bool `json:"can-lock"`
					CanReadSettings   bool `json:"can-read-settings"`
				} `json:"permissions"`
				Actions struct {
					IsDestroyable bool `json:"is-destroyable"`
				} `json:"actions"`
			} `json:"attributes"`
			Relationships struct {
				Organization struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
				} `json:"organization"`
				LatestRun struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
					Links struct {
						Related string `json:"related"`
					} `json:"links"`
				} `json:"latest-run"`
				CurrentRun struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
					Links struct {
						Related string `json:"related"`
					} `json:"links"`
				} `json:"current-run"`
				CurrentStateVersion struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
					Links struct {
						Related string `json:"related"`
					} `json:"links"`
				} `json:"current-state-version"`
			} `json:"relationships"`
			Links struct {
				Self string `json:"self"`
			} `json:"links"`
		} `json:"data"`
		Links struct {
			Self  string      `json:"self"`
			First string      `json:"first"`
			Prev  interface{} `json:"prev"`
			Next  interface{} `json:"next"`
			Last  string      `json:"last"`
		} `json:"links"`
		Meta struct {
			StatusCounts struct {
				Pending          int `json:"pending"`
				Planning         int `json:"planning"`
				Planned          int `json:"planned"`
				Confirmed        int `json:"confirmed"`
				Applying         int `json:"applying"`
				Applied          int `json:"applied"`
				Discarded        int `json:"discarded"`
				Errored          int `json:"errored"`
				Canceled         int `json:"canceled"`
				PolicyChecking   int `json:"policy-checking"`
				PolicyOverride   int `json:"policy-override"`
				PolicyChecked    int `json:"policy-checked"`
				PlanOnlyFinished int `json:"plan-only-finished"`
				None             int `json:"none"`
				Total            int `json:"total"`
			} `json:"status-counts"`
			Pagination struct {
				CurrentPage int         `json:"current-page"`
				PrevPage    interface{} `json:"prev-page"`
				NextPage    interface{} `json:"next-page"`
				TotalPages  int         `json:"total-pages"`
				TotalCount  int         `json:"total-count"`
			} `json:"pagination"`
		} `json:"meta"`
	}

	LatestRunInfo struct {
		Data []struct {
			ID         string `json:"id"`
			Type       string `json:"type"`
			Attributes struct {
				IsDestroy        bool   `json:"is-destroy"`
				Message          string `json:"message"`
				Source           string `json:"source"`
				Status           string `json:"status"`
				StatusTimestamps struct {
					AppliedAt   time.Time `json:"applied-at"`
					PlannedAt   time.Time `json:"planned-at"`
					ApplyingAt  time.Time `json:"applying-at"`
					PlanningAt  time.Time `json:"planning-at"`
					ConfirmedAt time.Time `json:"confirmed-at"`
				} `json:"status-timestamps"`
				CreatedAt  time.Time `json:"created-at"`
				HasChanges bool      `json:"has-changes"`
				Actions    struct {
					IsCancelable  bool `json:"is-cancelable"`
					IsConfirmable bool `json:"is-confirmable"`
					IsDiscardable bool `json:"is-discardable"`
				} `json:"actions"`
				PlanOnly    bool `json:"plan-only"`
				Permissions struct {
					CanApply        bool `json:"can-apply"`
					CanCancel       bool `json:"can-cancel"`
					CanDiscard      bool `json:"can-discard"`
					CanForceExecute bool `json:"can-force-execute"`
				} `json:"permissions"`
			} `json:"attributes"`
			Relationships struct {
				Workspace struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
				} `json:"workspace"`
				Apply struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
					Links struct {
						Related string `json:"related"`
					} `json:"links"`
				} `json:"apply"`
				ConfigurationVersion struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
					Links struct {
						Related string `json:"related"`
					} `json:"links"`
				} `json:"configuration-version"`
				ConfirmedBy struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
					Links struct {
						Related string `json:"related"`
					} `json:"links"`
				} `json:"confirmed-by"`
				CreatedBy struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
					Links struct {
						Related string `json:"related"`
					} `json:"links"`
				} `json:"created-by"`
				Plan struct {
					Data struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
					Links struct {
						Related string `json:"related"`
					} `json:"links"`
				} `json:"plan"`
				RunEvents struct {
					Data []struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
					Links struct {
						Related string `json:"related"`
					} `json:"links"`
				} `json:"run-events"`
				PolicyChecks struct {
					Data  []interface{} `json:"data"`
					Links struct {
						Related string `json:"related"`
					} `json:"links"`
				} `json:"policy-checks"`
				Comments struct {
					Data  []interface{} `json:"data"`
					Links struct {
						Related string `json:"related"`
					} `json:"links"`
				} `json:"comments"`
			} `json:"relationships"`
			Links struct {
				Self string `json:"self"`
			} `json:"links"`
		} `json:"data"`
		Included []struct {
			ID         string `json:"id"`
			Type       string `json:"type"`
			Attributes struct {
				Status           string `json:"status"`
				StatusTimestamps struct {
					QueuedAt   time.Time `json:"queued-at"`
					StartedAt  time.Time `json:"started-at"`
					FinishedAt time.Time `json:"finished-at"`
				} `json:"status-timestamps"`
				LogReadURL           string `json:"log-read-url"`
				ResourceAdditions    int    `json:"resource-additions"`
				ResourceChanges      int    `json:"resource-changes"`
				ResourceDestructions int    `json:"resource-destructions"`
			} `json:"attributes"`
			Relationships struct {
				StateVersions struct {
					Data []struct {
						ID   string `json:"id"`
						Type string `json:"type"`
					} `json:"data"`
				} `json:"state-versions"`
			} `json:"relationships"`
			Links struct {
				Self string `json:"self"`
			} `json:"links"`
		} `json:"included"`
		Links struct {
			Self  string      `json:"self"`
			First string      `json:"first"`
			Prev  interface{} `json:"prev"`
			Next  interface{} `json:"next"`
			Last  string      `json:"last"`
		} `json:"links"`
		Meta struct {
			Pagination struct {
				CurrentPage int         `json:"current-page"`
				PrevPage    interface{} `json:"prev-page"`
				NextPage    interface{} `json:"next-page"`
				TotalPages  int         `json:"total-pages"`
				TotalCount  int         `json:"total-count"`
			} `json:"pagination"`
		} `json:"meta"`
	}
)

const TFEBaseURL = "https://app.terraform.io/api/v2"

func NewTFE(organization, workspace, token string) (result *TFE) {
	result = &TFE{organization: organization, workspaceName: workspace, authorizationToken: token}
	return
}

func (tfe *TFE) Destroy() {
	tfe.authorizationToken = ""
	tfe.organization = ""
	tfe.workspaceName = ""
	tfe.id = ""
	tfe.request = nil
}

func (tfe *TFE) ChangeWorkspace(workspace string) {
	tfe.workspaceName = workspace
	tfe.id = "" // invalidate the cached data
}

func (tfe *TFE) getToken() (result string) {
	result = fmt.Sprintf("Bearer %s", tfe.authorizationToken)
	return
}

func (tfe *TFE) InitRequestHeaders(url string) (err error) {
	tfe.request, err = http.NewRequest("GET", url, nil)
	if err == nil {
		authToken := tfe.getToken()
		tfe.request.Header.Set("Authorization", authToken)
		tfe.request.Header.Set("Content-Type", "application/vnd.api+json")
	}
	return
}

func (tfe *TFE) getURL(url string) (result []byte, err error) {
	err = tfe.InitRequestHeaders(url)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(tfe.request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (tfe *TFE) listWorkspacesBytes() (result []byte, err error) {
	url := fmt.Sprintf("%s/organizations/%s/workspaces", TFEBaseURL, tfe.organization)

	return tfe.getURL(url)
}

func (tfe *TFE) getWorkspaceBytes() (result []byte, err error) {
	url := fmt.Sprintf("%s/organizations/%s/workspaces/%s", TFEBaseURL, tfe.organization, tfe.workspaceName)
	return tfe.getURL(url)
}

// GetWorkspaceID returns the cached workspace ID
func (tfe *TFE) GetWorkspaceID() string {
	return tfe.id
}

// WorkspaceID returns the ID for a workspace by making an API call to TFE
func (tfe *TFE) WorkspaceID() (workspaceID string, err error) {
	if tfe.id == "" {
		var responseBytes []byte
		responseBytes, err = tfe.getWorkspaceBytes()
		if err != nil {
			return
		}
		var gwr GetWorkspaceResponse
		err = json.Unmarshal(responseBytes, &gwr)
		if err != nil {
			fmt.Printf("%s\n", string(responseBytes))
			fmt.Printf("%v\n", err)
			return
		}
		if gwr.Data.ID == "" {
			var e ErrorResponse
			err = json.Unmarshal(responseBytes, &e)
			if err == nil {
				errMsg := fmt.Sprintf("Status: %s, Msg: %s", e.Errors[0].Status, e.Errors[0].Title)
				err = errors.New(errMsg)
				return "", err
			}
		}
		tfe.id = gwr.Data.ID
	}
	return tfe.id, nil
}

func (tfe *TFE) listWorkspaceRunsBytes() ([]byte, error) {
	var (
		workspaceID string
		err         error
	)
	if tfe.id == "" {
		workspaceID, err = tfe.WorkspaceID()
	} else {
		workspaceID = tfe.GetWorkspaceID()
	}
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/workspaces/%s/runs?include=apply", TFEBaseURL, workspaceID)
	// err = tfe.InitRequestHeaders(url)
	// if err != nil {
	// 	return nil, err
	// }
	// resp, err := http.DefaultClient.Do(tfe.request)
	// if err != nil {
	// 	return nil, err
	// }
	// defer resp.Body.Close()
	// return ioutil.ReadAll(resp.Body)
	return tfe.getURL(url)
}

// GetWorkspaceOutput gets the latest workspace run
func (tfe *TFE) GetWorkspaceOutput() (result []byte, err error) {
	var runbytes []byte
	runbytes, err = tfe.listWorkspaceRunsBytes()
	var latestRunInfo LatestRunInfo
	if err != nil {
		return
	}
	err = json.Unmarshal(runbytes, &latestRunInfo)
	if err != nil {
		return
	}
	url := latestRunInfo.Included[0].Attributes.LogReadURL

	runbytes, err = tfe.getURL(url)
	if err == nil {
		runOutput := string(runbytes)
		lines := strings.Split(runOutput, "\n")
		var (
			startWrite bool
			output     string
		)

		// look for lines appearing after Outputs:
		// then copy all of those lines
		ANSIEscape := string(rune(27)) + "[0m"
		for _, line := range lines {
			if startWrite {
				if line != "" {
					if strings.Contains(line, ANSIEscape) {
						// remove ANSI escape sequence
						line = strings.Replace(line, ANSIEscape, "", -1)
					}
					output = output + line + "\n"
				}
			} else {
				startWrite = startWrite || strings.HasPrefix(line, "Outputs:")
			}
		}

		result = []byte(output)

	}

	return
}

func (tfe *TFE) ListWorkspaceRuns() (result []byte, err error) {
	result, err = tfe.listWorkspaceRunsBytes()
	return
}
