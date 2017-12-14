package notifiers

import (
	"context"
	"testing"

	"github.com/grafana/grafana/pkg/components/null"
	"github.com/grafana/grafana/pkg/components/simplejson"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/alerting"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAlertmanagerNotifier(t *testing.T) {
	Convey("Alertmanager notifier tests", t, func() {

		Convey("Parsing alert notification from settings", func() {
			Convey("empty settings should return error", func() {
				json := `{ }`

				settingsJSON, _ := simplejson.NewJson([]byte(json))
				model := &m.AlertNotification{
					Name:     "alertmanager",
					Type:     "alertmanager",
					Settings: settingsJSON,
				}

				_, err := NewAlertmanagerNotifier(model)
				So(err, ShouldNotBeNil)
			})

			Convey("from settings", func() {
				json := `{ "url": "http://127.0.0.1:9093/" }`

				settingsJSON, _ := simplejson.NewJson([]byte(json))
				model := &m.AlertNotification{
					Name:     "alertmanager",
					Type:     "alertmanager",
					Settings: settingsJSON,
				}

				not, err := NewAlertmanagerNotifier(model)
				alertmanagerNotifier := not.(*AlertmanagerNotifier)

				So(err, ShouldBeNil)
				So(alertmanagerNotifier.Url, ShouldEqual, "http://127.0.0.1:9093/")
			})
		})

		Convey("Formatting alert notification", func() {
			Convey("Should correctly parse labels from message and evalMatch", func() {
				context := alerting.NewEvalContext(context.TODO(), &alerting.Rule{
					Name: "test_alert",
					Message: "A great description\n" +
						"With some details\n" +
						"\"label1\":\"value1\"\n" +
						"\"label2\":\"value2\"\n" +
						"\"label3\":\"value3\"\n",
				})
				match := alerting.EvalMatch{
					Metric: "fake.metric",
					Tags:   map[string]string{"tag1": "tagvalue1"},
				}
				expectedLabels := map[string]string{
					"alertname": "test_alert",
					"label1":    "value1",
					"label2":    "value2",
					"label3":    "value3",
					"tag1":      "tagvalue1",
				}
				actualLabels := parseLabels(context, &match)
				So(len(actualLabels), ShouldEqual, len(expectedLabels))
				for k, v := range expectedLabels {
					So(actualLabels[k], ShouldEqual, v)
				}
			})

			Convey("Should correctly add a 'metric' label if there is no tags", func() {
				context := alerting.NewEvalContext(context.TODO(), &alerting.Rule{
					Name:    "test_alert",
					Message: "A great description\n",
				})
				match := alerting.EvalMatch{
					Metric: "fake.metric",
					Tags:   map[string]string{},
				}
				expectedLabels := map[string]string{
					"alertname": "test_alert",
					"metric":    "fake.metric",
				}
				actualLabels := parseLabels(context, &match)
				So(len(actualLabels), ShouldEqual, len(expectedLabels))
				for k, v := range expectedLabels {
					So(actualLabels[k], ShouldEqual, v)
				}
			})

			Convey("Should correctly annotations", func() {
				context := alerting.NewEvalContext(context.TODO(), &alerting.Rule{
					Message: "A great description",
				})
				context.EvalMatches = append(context.EvalMatches,
					&alerting.EvalMatch{Value: null.FloatFrom(18.2), Metric: "foobar"})
				expectedAnnotations := map[string]string{
					"description": "A great description",
				}
				actualAnnotations := parseAnnotations(context)
				So(len(actualAnnotations), ShouldEqual, len(expectedAnnotations))
				for k, v := range expectedAnnotations {
					So(actualAnnotations[k], ShouldEqual, v)
				}
			})
		})
	})
}
