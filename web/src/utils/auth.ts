import { get as getCookie } from 'es-cookie';
import router from '@/router';

const COOKIE_NAME = 'sessionid';

/**
 * The token currently stored in the cookie.
 */
export function GetToken(): string {
  return getCookie(COOKIE_NAME) || '';
}

/**
 * `true` if a token is currently stored.
 */
export function IsAuthenticated(): boolean {
  return !!GetToken();
}

/**
 * Fails when an authorization or authentication error status is given.  If validation fails
 * users will be redirected to the login page.
 * @param status the HTTP status code of the response.
 */
export function ValidateStatus(status: number): boolean {
  const rval = status !== 401;

  if (!rval) {
    router.push({ name: 'login' });
  }

  return rval;
}
