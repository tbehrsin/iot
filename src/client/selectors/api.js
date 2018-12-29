
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
  api => memoize(key => {
    const body = api.getIn(['requests', key, 'body']);

    if (body) {
      return body.toJS();
    }
  })
);

export const url = createSelector(
  domain,
  api => memoize(key => {
    const url = api.getIn(['requests', key, 'url']);

    if (url) {
      return url;
    }
  })
);

export const error = createSelector(
  domain,
  api => memoize(key => {
    const error = api.getIn(['requests', key, 'error']);

    if (error) {
      return error.toJS();
    }
  })
);

export const connectionState = createSelector(
  domain,
  api => api.getIn(['connection', 'state'])
);
