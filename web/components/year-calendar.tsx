import { buildMonthGrid } from "@/lib/calendar";
import { VacationDots } from "@/components/vacation-dots";
import { Vacation } from "@/lib/types";

const MONTH_NAMES = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December"
];

type YearCalendarProps = {
  year: number;
  vacations: Vacation[];
  highlightedMemberId?: string | null;
};

export function YearCalendar({ year, vacations, highlightedMemberId }: YearCalendarProps) {
  return (
    <section className="year-grid">
      {MONTH_NAMES.map((monthName, monthIndex) => {
        const days = buildMonthGrid(year, monthIndex, vacations);
        return (
          <article key={monthName} className="month">
            <h3>{monthName}</h3>
            <div className="day-grid">
              {days.map((day) => {
                return (
                  <div
                    key={day.date.toISOString()}
                    className={`day ${day.inCurrentMonth ? "" : "day-muted"} ${
                      highlightedMemberId && day.vacationMemberIds?.includes(highlightedMemberId) ? "day-highlighted" : ""
                    }`}
                    title={day.vacationNames ? day.vacationNames.join(", ") : undefined}
                  >
                    {day.day}
                    <VacationDots colors={day.vacationColors} />
                  </div>
                );
              })}
            </div>
          </article>
        );
      })}
    </section>
  );
}
