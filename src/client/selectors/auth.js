
import { createSelector } from 'reselect';

export const domain = state => state.auth;

export const isPairing = createSelector(
  domain,
  auth => !!auth.get('isPairing')
);

export const isAuthenticating = createSelector(
  domain,
  auth => !!auth.get('isAuthenticating')
);

export const hasConnection = createSelector(
  domain,
  auth => !!auth.get('hasConnection')
);

export const error = createSelector(
  domain,
  auth => auth.get('error')
);

export const getToken = createSelector(
  domain,
  auth => auth.get('token')
);

export const hasToken = createSelector(
  getToken,
  token => !!token
);

export const getSeed = createSelector(
  domain,
  auth => auth.get('seed')
);

export const hasSeed = createSelector(
  getSeed,
  seed => !!seed
);

export const hasPinCode = createSelector(
  domain,
  auth => !!auth.get('hasPinCode')
);
