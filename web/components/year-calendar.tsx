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

const WEEKDAY_NAMES = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

type YearCalendarProps = {
  year: number;
  vacations: Vacation[];
  highlightedMemberIds?: string[];
};

export function YearCalendar({ year, vacations, highlightedMemberIds = [] }: YearCalendarProps) {
  const today = new Date();
  const todayYear = today.getUTCFullYear();
  const todayMonth = today.getUTCMonth();
  const todayDay = today.getUTCDate();

  return (
    <section className="year-grid">
      {MONTH_NAMES.map((monthName, monthIndex) => {
        const days = buildMonthGrid(year, monthIndex, vacations);
        return (
          <article key={monthName} className="month">
            <h3>{monthName}</h3>
            <div className="weekday-grid" aria-hidden="true">
              {WEEKDAY_NAMES.map((weekday) => (
                <div key={weekday} className="weekday-label">
                  {weekday}
                </div>
              ))}
            </div>
            <div className="day-grid">
              {days.map((day) => {
                const isToday =
                  day.inCurrentMonth &&
                  day.date.getUTCFullYear() === todayYear &&
                  day.date.getUTCMonth() === todayMonth &&
                  day.date.getUTCDate() === todayDay;

                return (
                  <div
                    key={day.date.toISOString()}
                    className={`day ${day.inCurrentMonth ? "" : "day-muted"} ${
                      highlightedMemberIds.length > 0 &&
                      day.vacationMemberIds?.some((memberId) => highlightedMemberIds.includes(memberId))
                        ? "day-highlighted"
                        : ""
                    } ${isToday ? "day-today" : ""}`}
                    title={day.vacationNames ? day.vacationNames.join(", ") : undefined}
                  >
                    {day.day}
                    <VacationDots dots={day.vacationDots} />
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
