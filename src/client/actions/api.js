
import { ws } from '../services';
import {
  API_REQUEST_BEGIN,
  API_REQUEST_END,
  API_REQUEST_ERROR,

  API_CONNECT_BEGIN,
  API_CONNECT_END,
  API_CONNECT_ERROR,

  API_DISCONNECT_BEGIN,
  API_DISCONNECT_END,
  API_DISCONNECT_ERROR,

  API_STATE_CONNECTED
} from '../constants';
import * as selectors from '../selectors';

export const request = (key, { method = 'GET', path = '/', query = {}, body = null }) => async (dispatch, getState) => {
  dispatch({
    type: API_REQUEST_BEGIN,
    payload: { key }
  });

  const token = selectors.auth.getToken(getState());

  try {
    console.info(`https://z3js.net${path}`, method, body);
    const response = await fetch(`https://z3js.net${path}`, {
      method,
      headers: {
        'X-Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: body == null ? null : JSON.stringify(body)
    });
    console.info(response);
    const json = await response.json();

    if (json.error) {
      const error = new ResourceError(json.error);
      dispatch({
        type: API_REQUEST_ERROR,
        payload: { key, error }
      });
      return;
    }

    dispatch({
      type: API_REQUEST_END,
      payload: { key, body: json.body }
    });
  } catch (error) {
    error = new ResourceError(error);
    dispatch({
      type: API_REQUEST_ERROR,
      payload: { key, error }
    });
    return;
  }
};

export const connect = () => async (dispatch, getState) => {
  dispatch({
    type: API_CONNECT_BEGIN
  });

  const token = selectors.auth.getToken(getState());

  try {
    const response = await fetch(`https://z3js.net/`, {
      method: 'OPTIONS',
      headers: {
        'X-Authorization': `Bearer ${token}`
      }
    });

    const location = response.headers.get('Location').replace('https://', 'wss://');
    await ws.connect(`${location}${token}`);

    dispatch({
      type: API_CONNECT_END
    });
  } catch (error) {
    dispatch({
      type: API_CONNECT_ERROR,
      payload: { error }
    });
  }
};

export const disconnect = () => async (dispatch, getState) => {
  if (selectors.api.connectionState(getState()) !== API_STATE_CONNECTED) {
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
    console.error(error);
    dispatch({
      type: API_DISCONNECT_ERROR,
      payload: { error }
    });
  }
}
