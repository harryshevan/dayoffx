import { Vacation } from "./types";

export type VacationDot = {
  colorHex: string;
  status: Vacation["status"];
};

export type DayVacation = {
  memberId: string;
  displayName: string;
  colorHex: string;
  status: Vacation["status"];
  fromDate: string;
  toDate: string;
};

export type DayCell = {
  day: number;
  date: Date;
  inCurrentMonth: boolean;
  vacationDots?: VacationDot[];
  vacationNames?: string[];
  vacationMemberIds?: string[];
  vacations?: DayVacation[];
};

const WEEK_START_MONDAY = 1;

function toIsoDateString(value: Date): string {
  const year = value.getUTCFullYear();
  const month = `${value.getUTCMonth() + 1}`.padStart(2, "0");
  const day = `${value.getUTCDate()}`.padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function buildMonthGrid(year: number, monthIndex: number, vacations: Vacation[]): DayCell[] {
  const firstDayOfMonth = new Date(Date.UTC(year, monthIndex, 1));
  const firstWeekday = (firstDayOfMonth.getUTCDay() + 7 - WEEK_START_MONDAY) % 7;
  const gridStart = new Date(firstDayOfMonth);
  gridStart.setUTCDate(gridStart.getUTCDate() - firstWeekday);

  return Array.from({ length: 42 }, (_, index) => {
    const date = new Date(gridStart);
    date.setUTCDate(gridStart.getUTCDate() + index);
    const dateString = toIsoDateString(date);

    const vacationsForDay = vacations.filter((item) => dateString >= item.fromDate && dateString <= item.toDate);
    return {
      day: date.getUTCDate(),
      date,
      inCurrentMonth: date.getUTCMonth() === monthIndex,
      vacationDots:
        vacationsForDay.length > 0
          ? vacationsForDay.map((item) => ({ colorHex: item.colorHex, status: item.status }))
          : undefined,
      vacationNames: vacationsForDay.length > 0 ? vacationsForDay.map((item) => item.displayName) : undefined,
      vacationMemberIds: vacationsForDay.length > 0 ? vacationsForDay.map((item) => item.memberId) : undefined,
      vacations:
        vacationsForDay.length > 0
          ? vacationsForDay.map((item) => ({
              memberId: item.memberId,
              displayName: item.displayName,
              colorHex: item.colorHex,
              status: item.status,
              fromDate: item.fromDate,
              toDate: item.toDate
            }))
          : undefined
    };
  });
}
