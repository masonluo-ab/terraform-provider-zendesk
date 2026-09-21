package zendesk

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nukosuke/go-zendesk/zendesk"
	"github.com/nukosuke/go-zendesk/zendesk/mock"
)

func TestMarshalTrigger(t *testing.T) {
	expected := zendesk.Trigger{
		Title:       "title",
		Description: "blabla",
		Active:      true,
	}
	m := &identifiableMapGetterSetter{
		mapGetterSetter: mapGetterSetter{},
	}

	err := marshalTrigger(expected, m)
	if err != nil {
		t.Fatalf("Failed to marshal map %v", err)
	}

	v, ok := m.GetOk("title")
	if !ok {
		t.Fatal("Failed to get title value")
	}
	if v != expected.Title {
		t.Fatalf("trigger had incorrect title value %v. should have been %v", v, expected.Title)
	}

	v, ok = m.GetOk("description")
	if !ok {
		t.Fatal("Failed to get description value")
	}
	if v != expected.Description {
		t.Fatalf("trigger had incorrect description value %v. should have been %v", v, expected.Description)
	}

	v, ok = m.GetOk("active")
	if !ok {
		t.Fatal("Failed to get active value")
	}
	if v != expected.Active {
		t.Fatalf("trigger had incorrect active value %v. should have been %v", v, expected.Active)
	}
}

func TestUnmarshalTrigger(t *testing.T) {
	m := &identifiableMapGetterSetter{
		id: "100",
		mapGetterSetter: mapGetterSetter{
			"title":       "Auto reply",
			"description": "reply automatically",
			"active":      true,
		},
	}

	trg, err := unmarshalTrigger(m)
	if err != nil {
		t.Fatalf("unmarshal returned an error: %v", err)
	}

	if v := m.Get("title"); trg.Title != v {
		t.Fatalf("trigger had title value %v. should have been %v", trg.Title, v)
	}
}

func TestMarshalTriggerValue(t *testing.T) {
	for _, tc := range []struct {
		name     string
		value    interface{}
		expected string
	}{
		{name: "string", value: "34", expected: "34"},
		{name: "list", value: []interface{}{"3926199", "3926239"}, expected: `["3926199","3926239"]`},
		{name: "absent", value: nil, expected: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, err := marshalTriggerValue(tc.value)
			if err != nil {
				t.Fatalf("marshalTriggerValue returned an error: %v", err)
			}
			if v != tc.expected {
				t.Fatalf("marshalTriggerValue returned %v. should have been %v", v, tc.expected)
			}
		})
	}
}

func TestUnmarshalTriggerValue(t *testing.T) {
	v, err := unmarshalTriggerValue("34")
	if err != nil {
		t.Fatalf("unmarshalTriggerValue returned an error: %v", err)
	}
	if v != "34" {
		t.Fatalf("unmarshalTriggerValue returned %v. should have been 34", v)
	}

	v, err = unmarshalTriggerValue(`["3926199","3926239"]`)
	if err != nil {
		t.Fatalf("unmarshalTriggerValue returned an error: %v", err)
	}
	list, ok := v.([]interface{})
	if !ok {
		t.Fatalf("unmarshalTriggerValue returned %T. should have been a list", v)
	}
	if len(list) != 2 || list[0] != "3926199" || list[1] != "3926239" {
		t.Fatalf("unmarshalTriggerValue returned %v. should have been [3926199 3926239]", list)
	}

	if _, err = unmarshalTriggerValue("[not json"); err == nil {
		t.Fatal("unmarshalTriggerValue accepted a value that is not a list")
	}
}

// A condition on a multi-select field, such as Zendesk's custom ticket statuses, has a list for its value.
func TestMarshalTriggerWithListValuedCondition(t *testing.T) {
	expected := zendesk.Trigger{Title: "title"}
	expected.Conditions.All = []zendesk.TriggerCondition{
		{
			Field:    "custom_status_id",
			Operator: "not_includes",
			Value:    []interface{}{"3926199", "3926239"},
		},
	}
	m := &identifiableMapGetterSetter{
		mapGetterSetter: mapGetterSetter{},
	}

	err := marshalTrigger(expected, m)
	if err != nil {
		t.Fatalf("Failed to marshal map %v", err)
	}

	v, ok := m.GetOk("all")
	if !ok {
		t.Fatal("Failed to get all value")
	}
	conditions, ok := v.([]map[string]interface{})
	if !ok {
		t.Fatalf("all had type %T. should have been a list of conditions", v)
	}
	if len(conditions) != 1 {
		t.Fatalf("all had %d conditions. should have been 1", len(conditions))
	}
	if value := conditions[0]["value"]; value != `["3926199","3926239"]` {
		t.Fatalf("condition had incorrect value %v. should have been the list's JSON encoding", value)
	}
}

func TestCreateTrigger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock.NewClient(ctrl)
	i := newIdentifiableGetterSetter()
	out := zendesk.Trigger{
		ID:    12345,
		Title: "trigger",
	}

	m.EXPECT().CreateTrigger(gomock.Any(), gomock.Any()).Return(out, nil)
	if diags := createTrigger(context.Background(), i, m); len(diags) != 0 {
		t.Fatal("CreateTrigger return an error")
	}

	if v := i.Id(); v != "12345" {
		t.Fatalf("CreateTrigger did not set resource id. Id was %s", v)
	}

	if v := i.Get("title"); v != "trigger" {
		t.Fatalf("CreateTrigger did not set resource title. title was %s", v)
	}
}

func TestReadTrigger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock.NewClient(ctrl)
	i := newIdentifiableGetterSetter()
	i.SetId("12345")

	expected := zendesk.Trigger{
		Title:  "trigger",
		Active: true,
	}
	m.EXPECT().GetTrigger(gomock.Any(), gomock.Eq(int64(12345))).Return(expected, nil)
	if diags := readTrigger(context.Background(), i, m); len(diags) != 0 {
		t.Fatalf("GetTrigger received an error when calling: %v", diags)
	}
}

func TestUpdateTrigger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock.NewClient(ctrl)
	i := &identifiableMapGetterSetter{
		id:              "12345",
		mapGetterSetter: mapGetterSetter{},
	}

	m.EXPECT().UpdateTrigger(gomock.Any(), gomock.Eq(int64(12345)), gomock.Any()).Return(zendesk.Trigger{}, nil)
	if diags := updateTrigger(context.Background(), i, m); len(diags) != 0 {
		t.Fatalf("updateTrigger returned an error %v", diags)
	}
}

func TestDeleteTrigger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mock.NewClient(ctrl)
	d := newIdentifiableGetterSetter()

	d.SetId("1234")

	c.EXPECT().DeleteTrigger(gomock.Any(), gomock.Eq(int64(1234))).Return(nil)
	diags := deleteTrigger(context.Background(), d, c)
	if len(diags) != 0 {
		t.Fatalf("Got error from resource delete: %v", diags)
	}
}

func testTriggerDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(zendesk.TriggerAPI)

	for k, r := range s.RootModule().Resources {
		if r.Type != "zendesk_trigger" {
			continue
		}

		id, err := atoi64(r.Primary.ID)
		if err != nil {
			return err
		}

		ctx := context.Background()
		_, err = client.GetTrigger(ctx, id)
		if err == nil {
			return fmt.Errorf("did not get error from zendesk when trying to fetch the destroyed trigger. resource name %s", k)
		}

		zdresp, ok := err.(zendesk.Error)
		if !ok {
			return fmt.Errorf("error %v cannot be asserted as a zendesk error", err)
		}

		if zdresp.Status() != http.StatusNotFound {
			return fmt.Errorf("did not get a not found error after destroy. error was %v", zdresp)
		}
	}
	return nil
}

func TestAccTriggerExample(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testTriggerDestroyed,
		Steps: []resource.TestStep{
			{
				Config: readExampleConfig(t, "resources/zendesk_trigger/resource.tf"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zendesk_trigger.auto-reply-trigger", "title", "Auto Reply Trigger"),
					resource.TestCheckResourceAttr("zendesk_trigger.auto-reply-trigger", "active", "true"),
					resource.TestCheckResourceAttrSet("zendesk_trigger.auto-reply-trigger", "all.#"),
					resource.TestCheckResourceAttrSet("zendesk_trigger.auto-reply-trigger", "action.#"),
				),
			},
		},
	})
}
