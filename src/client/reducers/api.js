
import { fromJS } from 'immutable';
import {
  API_REQUEST_BEGIN,
  API_REQUEST_END,
  API_REQUEST_ERROR,
  API_REQUEST_RESET,
  API_CONNECT_BEGIN,
  API_CONNECT_END,
  API_CONNECT_ERROR,
  API_DISCONNECT_BEGIN,
  API_DISCONNECT_END,
  API_DISCONNECT_ERROR,
  API_STATE_DISCONNECTED,
  API_STATE_CONNECTING,
  API_STATE_CONNECTED,
  API_STATE_DISCONNECTING,
  API_STATE_ERROR
} from '../constants';

const initialState = {
  requests: {},
  connection: {
    state: API_STATE_DISCONNECTED
  }
};

export default (state = fromJS(initialState), { type, payload }) => {
  switch(type) {
    case API_REQUEST_BEGIN: {
      const { key } = payload;
      return state.setIn(['requests', key], fromJS({}));
    }
    case API_REQUEST_END: {
      const { key, body } = payload;
      return state.setIn(['requests', key, 'body'], fromJS(body));
    }
    case API_REQUEST_ERROR: {
      const { key, error } = payload;
      return state.setIn(['requests', key, 'error'], fromJS(error));
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
    default:
      return state;
  }
};
