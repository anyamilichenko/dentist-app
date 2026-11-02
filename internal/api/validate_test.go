package api

import (
	"testing"
	"time"

	"dentist-app/internal/store"
)

func ptr[T any](v T) *T { return &v }

func TestNormalizePatient(t *testing.T) {
	in := store.PatientInput{
		LastName:   "  Иванов ",
		FirstName:  "Иван",
		MiddleName: ptr("   "),
		Phone:      ptr(" +7 (916) 123-45-67 "),
	}
	if err := normalizePatient(&in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.LastName != "Иванов" {
		t.Errorf("last name not trimmed: %q", in.LastName)
	}
	if in.MiddleName != nil {
		t.Errorf("blank middle name should become nil, got %q", *in.MiddleName)
	}
	if in.Phone == nil || *in.Phone != "+7 (916) 123-45-67" {
		t.Errorf("phone not trimmed: %v", in.Phone)
	}

	bad := []store.PatientInput{
		{FirstName: "Иван"},
		{LastName: "Иванов"},
		{LastName: "Иванов", FirstName: "Иван", Phone: ptr("позвонить")},
	}
	for _, in := range bad {
		if err := normalizePatient(&in); err == nil {
			t.Errorf("expected error for %+v", in)
		}
	}
}

func TestNormalizeService(t *testing.T) {
	ok := store.ServiceInput{Name: " Лечение кариеса ", DurationMinutes: 60, Price: 5000}
	if err := normalizeService(&ok); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok.Name != "Лечение кариеса" {
		t.Errorf("name not trimmed: %q", ok.Name)
	}

	bad := []store.ServiceInput{
		{Name: "", DurationMinutes: 30},
		{Name: "Осмотр", DurationMinutes: 0},
		{Name: "Осмотр", DurationMinutes: 30, Price: -1},
	}
	for _, in := range bad {
		if err := normalizeService(&in); err == nil {
			t.Errorf("expected error for %+v", in)
		}
	}
}

func TestNormalizeAppointment(t *testing.T) {
	start := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)
	ok := store.AppointmentInput{PatientID: 1, StartsAt: start, EndsAt: start.Add(30 * time.Minute)}
	if err := normalizeAppointment(&ok); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bad := []store.AppointmentInput{
		{StartsAt: start, EndsAt: start.Add(time.Hour)},
		{PatientID: 1, StartsAt: start, EndsAt: start},
		{PatientID: 1, StartsAt: start, EndsAt: start.Add(-time.Hour)},
		{PatientID: 1, EndsAt: start},
		{PatientID: 1, ServiceID: ptr(int64(0)), StartsAt: start, EndsAt: start.Add(time.Hour)},
	}
	for _, in := range bad {
		if err := normalizeAppointment(&in); err == nil {
			t.Errorf("expected error for %+v", in)
		}
	}
}
