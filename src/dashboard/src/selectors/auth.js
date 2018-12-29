
import { createSelector } from 'reselect';

export const domain = state => state.auth;

export const getToken = createSelector(
  domain,
  auth => auth.get('token')
);

export const hasToken = createSelector(
  getToken,
  token => !!token
);

export const gateway = createSelector(
  getToken,
  token => {
    if (!token) {
      return false;
    }

    let [,claims] = token.split(/\./g);
    claims = JSON.parse(atob(claims));

    return claims.gateway;
  }
);

export const hasAppToken = createSelector(
  getToken,
  token => {
    if (!token) {
      return false;
    }

    let [,claims] = token.split(/\./g);
    claims = JSON.parse(atob(claims));

    return claims.aud === 'app';
  }
);

export const hasEmailToken = createSelector(
  getToken,
  token => {
    if (!token) {
      return false;
    }

    let [,claims] = token.split(/\./g);
    claims = JSON.parse(atob(claims));

    return claims.aud === 'email';
  }
);
