
import { fromJS } from 'immutable';
import {
  API_SET_LOCAL,
  API_REQUEST_BEGIN,
  API_REQUEST_END,
  API_REQUEST_ERROR,

  API_SET_IN,

  API_REQUEST_RESET,

  API_CONNECT_BEGIN,
  API_CONNECT_END,
  API_CONNECT_ERROR,

  API_DISCONNECT_BEGIN,
  API_DISCONNECT_END,
  API_DISCONNECT_ERROR,

  API_RECONNECTING,

  API_STATE_REQUESTING,
  API_STATE_COMPLETE,

  API_STATE_DISCONNECTED,
  API_STATE_CONNECTING,
  API_STATE_CONNECTED,
  API_STATE_RECONNECTING,
  API_STATE_DISCONNECTING,
  API_STATE_ERROR
} from '../constants';
import { constants } from '../../app.json';

const initialState = {
  requests: {},
  local: !constants.forceServer
};

export default (state = fromJS(initialState), { type, payload }) => {
  switch(type) {
    case API_SET_LOCAL: {
      const { local } = payload;
      return state.set('local', !!local);
    }
    case API_REQUEST_BEGIN: {
      const { key } = payload;
      return state.setIn(['requests', key, 'state'], API_STATE_REQUESTING).deleteIn(['requests', key, 'refresh']);
    }
    case API_REQUEST_END: {
      const { key, body, url, refresh } = payload;
      return state.setIn(['requests', key, 'body'], fromJS(body)).setIn(['requests', key, 'state'], API_STATE_COMPLETE).setIn(['requests', key, 'url'], url).setIn(['requests', key, 'refresh'], fromJS(refresh));
    }
    case API_REQUEST_ERROR: {
      const { key, error } = payload;
      return state.setIn(['requests', key, 'error'], fromJS(error)).setIn(['requests', key, 'state'], API_STATE_ERROR).deleteIn(['requests', key, 'refresh']);
    }
    case API_SET_IN: {
      const { key, path, value } = payload;
      return state.setIn(['requests', key, 'body', ...path], fromJS(value));
    }
    case API_REQUEST_RESET: {
      const { key } = payload;
      return state.deleteIn(['requests', key]);
    }
    case API_CONNECT_BEGIN: {
      return state.setIn(['connection', 'state'], API_STATE_CONNECTING).deleteIn(['connection', 'error']);
    }
    case API_CONNECT_END: {
      return state.setIn(['connection', 'state'], API_STATE_CONNECTED).deleteIn(['connection', 'error']);
    }
    case API_CONNECT_ERROR: {
      const { error } = payload;
      return state.setIn(['connection', 'state'], API_STATE_ERROR).setIn(['connection', 'error'], fromJS(error));
    }
    case API_DISCONNECT_BEGIN: {
      return state.setIn(['connection', 'state'], API_STATE_DISCONNECTING).deleteIn(['connection', 'error']);
    }
    case API_DISCONNECT_END: {
      return state.setIn(['connection', 'state'], API_STATE_DISCONNECTED).deleteIn(['connection', 'error']);
    }
    case API_DISCONNECT_ERROR: {
      const { error } = payload;
      return state.setIn(['connection', 'state'], API_STATE_ERROR).setIn(['connection', 'error'], fromJS(error));
    }
    case API_RECONNECTING: {
      const { error } = payload;
      return state.setIn(['connection', 'state'], API_STATE_RECONNECTING).setIn(['connection', 'error'], fromJS(error));
    }
    default:
      return state;
  }
};
