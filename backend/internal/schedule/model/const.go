package model

const (
	CalendarPersonal   int16 = 1
	CalendarDepartment int16 = 2
	CalendarProject    int16 = 3

	MemberViewer int16 = 1
	MemberEditor int16 = 2

	EventConfirmed int16 = 0
	EventCancelled int16 = 1

	RecurrenceNone    = "none"
	RecurrenceDaily   = "daily"
	RecurrenceWeekly  = "weekly"
	RecurrenceMonthly = "monthly"

	AttendeePending   int16 = 0
	AttendeeAccepted  int16 = 1
	AttendeeDeclined  int16 = 2
	AttendeeTentative int16 = 3

	RemindApp int16 = 1

	SourcePersonal  = "personal"
	SourceTimetable = "timetable"
	SourceImport    = "import"
	SourceGoogle    = "google"

	OriginManual    = "manual"
	OriginGenerated = "generated"
	OriginICS       = "ics"
	OriginGoogle    = "google"
)

func ValidCalendarType(t int16) bool {
	return t >= CalendarPersonal && t <= CalendarProject
}

func ValidMemberRole(r int16) bool {
	return r == MemberViewer || r == MemberEditor
}

func ValidRecurrence(v string) bool {
	switch v {
	case "", RecurrenceNone, RecurrenceDaily, RecurrenceWeekly, RecurrenceMonthly:
		return true
	default:
		return false
	}
}

func NormalizeRecurrence(v string) string {
	if v == "" {
		return RecurrenceNone
	}
	return v
}

func ValidRemindMinutes(m int) bool {
	return m == 5 || m == 15 || m == 30 || m == 60
}

func ValidAttendeeResponse(s int16) bool {
	return s >= AttendeePending && s <= AttendeeTentative
}

func ValidSource(s string) bool {
	switch s {
	case "", SourcePersonal, SourceTimetable, SourceImport, SourceGoogle:
		return true
	default:
		return false
	}
}

func NormalizeSource(s string) string {
	if s == "" {
		return SourcePersonal
	}
	return s
}

func DefaultLayerColor(source string) string {
	switch NormalizeSource(source) {
	case SourceTimetable:
		return "#059669"
	case SourceImport:
		return "#d97706"
	case SourceGoogle:
		return "#4285f4"
	default:
		return "#2563eb"
	}
}

func IsPersonalLayer(calendarType int16, source string) bool {
	return calendarType == CalendarPersonal && NormalizeSource(source) == SourcePersonal
}
