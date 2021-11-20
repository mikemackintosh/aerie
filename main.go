package aerie

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/reports/v1"
	reports "google.golang.org/api/admin/reports/v1"
	"google.golang.org/api/option"
)

var srv = &admin.Service{}

func init() {
	var err error
	var serviceAccountJson []byte

	serviceAccountB64 := os.Getenv("SERVICE_ACCOUNT_JSON")
	if len(serviceAccountB64) != 0 {
		serviceAccountJson, err = base64.StdEncoding.DecodeString(serviceAccountB64)
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}
	} else {
		serviceAccountJson, err = ioutil.ReadFile("service-account.json")
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}
	}

	config, err := google.JWTConfigFromJSON(serviceAccountJson, reports.AdminReportsAuditReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	config.Subject = os.Getenv("SERVICE_ACCOUNT_EMAIL")

	var ctx = context.Background()
	client := config.Client(ctx)

	srv, err = reports.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve reports Client %v", err)
	}
}

func StartWatching(user, application string, eventName []string) error {
	for _, event := range eventName {
		var activitySvc = reports.NewActivitiesService(srv)

		uuid, err := uuid.NewUUID()
		if err != nil {
			log.Fatalf("Unable to cache oauth token: %v", err)
		}

		expiration := time.Now()
		expiration = expiration.Add(time.Hour * 5)

		// Creates a new channel resource with an experi
		channel := reports.Channel{
			Id:         uuid.String(),
			Payload:    true,
			Address:    os.Getenv("WEBHOOK_URL"),
			Type:       "web_hook",
			Token:      os.Getenv("SHARED_SECRET"),
			Expiration: expiration.Unix() * 1000,
		}
		req := activitySvc.Watch(user, application, &channel)

		if len(event) > 0 {
			req.EventName(event)
		}

		resp, err := req.Do()
		if err != nil {
			log.Fatalf("Unable to cache oauth token: %v", err)
		}

		fmt.Printf("resp: %+v\n", resp)
	}
	return nil
}

func GetReportEvents(user, application string) error {
	r, err := srv.Activities.List(user, application).MaxResults(100).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve logins to domain. %v", err)
	}

	if len(r.Items) == 0 {
		fmt.Println("No logins found.")
	} else {
		fmt.Println("Logins:")
		for _, a := range r.Items {
			t, err := time.Parse(time.RFC3339Nano, a.Id.Time)
			if err != nil {
				fmt.Println("Unable to parse login time.")
				// Set time to zero.
				t = time.Time{}
				return err
			}
			fmt.Printf("%s: %s %s\n", t.Format(time.RFC822), a.Actor.Email,
				a.Events[0].Name)
		}
	}

	return nil
}

func HandlerWorkspaceNotifier(c echo.Context) error {
	b, err := httputil.DumpRequest(c.Request(), true)
	if err != nil {
		return c.String(http.StatusBadRequest, "FAIL")
	}

	log.Printf("-> request: %#v\n", string(b))
	return c.String(http.StatusOK, "OK")
}
