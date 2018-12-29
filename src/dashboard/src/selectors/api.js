
import Immutable from 'immutable';
import { createSelector } from 'reselect';
import memoize from 'lodash.memoize';

import {
  API_STATE_REQUESTING,
  API_STATE_COMPLETE,
  API_STATE_ERROR
} from '../constants';

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

export const bodyIn = createSelector(
  domain,
  api => memoize((key, path) => {
    const value = api.getIn(['requests', key, 'body', ...path]);

    if (value && value.toJS) {
      return value.toJS();
    } else {
      return value;
    }
  })
)

export const findInBody = createSelector(
  domain,
  api => memoize((key, callback) => {
    const body = api.getIn(['requests', key, 'body']);

    if (body) {
      const found = body.find(callback);
      if (found) {
        return found.toJS();
      }
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

export const local = createSelector(
  domain,
  api => !!api.get('local')
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

export const requestState = createSelector(
  domain,
  api => memoize(key => api.getIn(['requests', key, 'state']))
);

export const requestStateAll = createSelector(
  domain,
  api => {
    const error = !!api.get('requests').find(r => r.get('state') === API_STATE_ERROR);
    if (error) {
      return API_STATE_ERROR;
    }

    const requesting = !!api.get('requests').find(r => r.get('state') === API_STATE_REQUESTING);
    if (requesting) {
      return API_STATE_REQUESTING;
    }

    return API_STATE_COMPLETE;
  }
);

export const refreshRequests = createSelector(
  domain,
  api => api.get('requests').filter(r => r.get('refresh')).map(request => request.get('refresh')).toJS()
)
