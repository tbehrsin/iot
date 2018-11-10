
import { createSelector } from 'reselect';

export const selectAuthDomain = state => state.auth;

export const selectAuthLoggedIn = createSelector(
  selectAuthDomain,
  auth => auth.get('loggedIn')
);
