
import {
  API_SET_LOCAL,

  API_REQUEST_BEGIN,
  API_REQUEST_END,
  API_REQUEST_ERROR,

  API_REQUEST_RESET,

  API_SET_IN,

  API_CONNECT_BEGIN,
  API_CONNECT_END,
  API_CONNECT_ERROR,

  API_DISCONNECT_BEGIN,
  API_DISCONNECT_END,
  API_DISCONNECT_ERROR,

  API_RECONNECTING,

  API_STATE_RECONNECTING,
  API_STATE_CONNECTED,
  API_STATE_ERROR
} from '../constants';

import { constants as config } from 'config';

import * as selectors from '../selectors';
import { ws } from '../services';

export const setLocal = (local) => ({
  type: API_SET_LOCAL,
  payload: { local }
});

export const reset = (key) => ({
  type: API_REQUEST_RESET,
  payload: { key }
});

export const request = (key, options) => async (dispatch, getState) => {
  const { method = 'GET', path = '/', query = {}, body = null, retry = true, refresh = method === 'GET' } = options;

  dispatch({
    type: API_REQUEST_BEGIN,
    payload: { key }
  });

  const local = selectors.api.local(getState());
  const token = selectors.auth.getToken(getState());
  const gateway = selectors.auth.gateway(getState());

  try {
    const response = await fetch(`https://${local ? 'local.' : ''}${gateway}.${config.domain}${path}`, {
      method,
      headers: {
        'X-Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: body == null ? null : JSON.stringify(body)
    });
    let json;
    try {
      json = await response.json();
    } catch (error) {
      json = {};
    }

    if (json.error) {
      const error = new ResourceError(json.error);
      dispatch({
        type: API_REQUEST_ERROR,
        payload: { key, error }
      });
      return;
    }

    // strip out any unneeded port as this is likely to cause problems with other user agents
    let url = response.url;
    url = url.replace(/^(https:\/\/[^\/]+):443\//, (g, g1) => `${g1}/`)
    url = url.replace(/^(http:\/\/[^\/]+):80\//, (g, g1) => `${g1}/`)

    dispatch({
      type: API_REQUEST_END,
      payload: { key, body: json.body, url, refresh: refresh ? options : false }
    });
  } catch (error) {
    error = new ResourceError(error);
    dispatch({
      type: API_REQUEST_ERROR,
      payload: { key, error }
    });

    if (retry) {
      setTimeout(() => {

        dispatch(request(key, options));
      }, 5000);
    }
    return;
  }
};

export const setIn = (key, path, value) => ({
  type: API_SET_IN,
  payload: {
    key,
    path,
    value
  }
});


export const connect = () => async (dispatch, getState) => {
  dispatch({
    type: API_CONNECT_BEGIN
  });

  const local = selectors.api.local(getState());
  const token = selectors.auth.getToken(getState());
  const gateway = selectors.auth.gateway(getState());

  try {
    const location = `wss://${local ? 'local.' : ''}${gateway}.${config.domain}/`;
    console.info("Connecting", location);
    await ws.connect(`${location}${token}`);
  } catch (error) {
    dispatch({
      type: API_CONNECT_ERROR,
      payload: { error }
    });

    setTimeout(() => {
      if (selectors.api.connectionState(getState()) === API_STATE_ERROR) {
        dispatch(connect());
      }
    }, 1000);
  }
};

export const disconnect = () => async (dispatch, getState) => {
  if ([API_STATE_CONNECTED, API_STATE_RECONNECTING].indexOf(selectors.api.connectionState(getState())) !== -1) {
    return;
  }

  dispatch({
    type: API_DISCONNECT_BEGIN
  });

  try {
    await ws.disconnect();

    dispatch({
      type: API_DISCONNECT_END
    });
  } catch (error) {
    dispatch({
      type: API_DISCONNECT_ERROR,
      payload: { error }
    });
  }
}

export const reconnecting = (error) => async (dispatch, getState) => {
  dispatch({
    type: API_RECONNECTING,
    payload: { error }
  });

  setTimeout(() => {
    if (selectors.api.connectionState(getState()) === API_STATE_RECONNECTING) {
      dispatch(connect());
    }
  }, 1000);
};

export const connected = () => ({
  type: API_CONNECT_END
});
