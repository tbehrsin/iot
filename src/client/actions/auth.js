
import { Buffer } from 'buffer';
import { sha256 } from 'react-native-sha256';
import { generateSecureRandom } from 'react-native-securerandom';

import { ble } from '../services';
import {
  AUTH_INITIALIZE,
  AUTH_SET_SEED,
  AUTH_SET_TOKEN,
  AUTH_AUTHENTICATE_BEGIN,
  AUTH_AUTHENTICATE_END,
  AUTH_AUTHENTICATE_ERROR,
  AUTH_PAIR_BEGIN,
  AUTH_PAIR_END,
  AUTH_PAIR_ERROR
} from '../constants';

import { history } from '../router';
import * as api from './api';
import * as selectors from '../selectors';

export const initialize = () => ({
  type: AUTH_INITIALIZE
});

export const pair = () => async (dispatch) => {
  dispatch({
    type: AUTH_PAIR_BEGIN
  });

  try {
    console.debug('starting bluetooth');
    await ble.start();

    console.debug('connecting to gateway');
    await ble.device.connect();
  } catch (error) {
    console.error(error);
    dispatch({
      type: AUTH_PAIR_ERROR,
      payload: { error }
    });
    return;
  }

  try {
    const { body } = await ble.connection.send({
      type: 'auth/GET_PIN_CODE_SEED',
      payload: {}
    });
    dispatch(setSeed(body.seed));
  } catch (error) {
    // if a pin code does not already exist then it won't have a seed either and we can also assume that we'll need to
    // initialize a new gateway
    if (error.code === ResourceError.NotFound) {
      const { body } = await ble.connection.send({
        type: 'gateway/CREATE_GATEWAY',
        payload: {}
      });
      const seed = Buffer.from(await generateSecureRandom(20)).toString('base64');
      dispatch(setSeed(seed, false));
    } else {
      console.error(error);
      dispatch({
        type: AUTH_PAIR_ERROR,
        payload: { error }
      });
      return;
    }
  }

  dispatch({
    type: AUTH_PAIR_END
  });
};

export const setSeed = (seed, hasPinCode = true) => ({
  type: AUTH_SET_SEED,
  payload: {
    seed,
    hasPinCode
  }
});

export const authenticate = (pin) => async (dispatch, getState) => {
  dispatch({
    type: AUTH_AUTHENTICATE_BEGIN
  });

  let state = getState();

  if (!selectors.auth.hasSeed(state)) {
    dispatch({
      type: AUTH_AUTHENTICATE_ERROR,
      payload: { error: new Error('tried to authenticate without a seed') }
    });
    return;
  }
  const seedString = selectors.auth.getSeed(state);
  const seed = Buffer.from(seedString, 'base64').toString();
  const hash = Buffer.from(await sha256(`${pin}${seed}`), 'hex').toString('base64');

  try {
    if (selectors.auth.hasPinCode(state)) {
       const { body } = await ble.connection.send({
         type: 'auth/VERIFY_PIN_CODE',
         payload: { hash }
       });
       const { token } = body;
       dispatch(setToken(token));
    } else {
      const { body } = await ble.connection.send({
        type: 'auth/SET_PIN_CODE',
        payload: { seed: seedString, hash }
      });
      const { token } = body;
      dispatch(setToken(token));
    }
  } catch (error) {
    console.error(error);
    dispatch({
      type: AUTH_AUTHENTICATE_ERROR,
      payload: { error }
    });
    return;
  }

  dispatch({
    type: AUTH_AUTHENTICATE_END
  });

  history.go(-history.entries.length);
  history.push('/');
};

export const setToken = (token) => ({
  type: AUTH_SET_TOKEN,
  payload: { token }
});
