
import { createSelector } from 'reselect';
import memoize from 'lodash.memoize';

export const domain = state => state.api;

export const isPending = createSelector(
  domain,
  api => memoize(key => !api.getIn(['requests', key, 'body']) && !api.getIn(['requests', key, 'error']))
);

export const isSuccess = createSelector(
  domain,
  api => memoize(key => !!api.getIn(['requests', key, 'body']))
);

export const isFail = createSelector(
  domain,
  api => memoize(key => !!api.getIn(['requests', key, 'error']))
);

export const body = createSelector(
  domain,
  api => memoize(key => api.getIn(['requests', key, 'body']))
);

export const error = createSelector(
  domain,
  api => memoize(key => api.getIn(['requests', key, 'error']))
);
