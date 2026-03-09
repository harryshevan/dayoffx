import createMiddleware from "next-intl/middleware";
import { NextRequest, NextResponse } from "next/server";
import { routing } from "@/i18n/routing";

const intlMiddleware = createMiddleware(routing);
const LOCALE_COOKIE = "NEXT_LOCALE";

export default function middleware(request: NextRequest) {
  const { pathname, searchParams } = request.nextUrl;
  const localeMatch = pathname.match(/^\/(en|ru)(\/.*)?$/);
  const cookieLocale = request.cookies.get(LOCALE_COOKIE)?.value;

  if (localeMatch && searchParams.get("onboarding") !== "1") {
    const url = request.nextUrl.clone();
    url.pathname = "/";
    const response = NextResponse.redirect(url);
    response.cookies.set(LOCALE_COOKIE, localeMatch[1], {
      path: "/",
      sameSite: "lax"
    });
    return response;
  }

  if (!cookieLocale && !localeMatch) {
    const url = request.nextUrl.clone();
    const response = NextResponse.redirect(url);
    response.cookies.set(LOCALE_COOKIE, routing.defaultLocale, {
      path: "/",
      sameSite: "lax"
    });
    return response;
  }

  return intlMiddleware(request);
}

export const config = {
  matcher: "/((?!api|_next|_vercel|.*\\..*).*)"
};
