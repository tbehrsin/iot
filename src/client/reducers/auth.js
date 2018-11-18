
import { fromJS } from 'immutable';
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

const initialState = {
};

export default (state = fromJS(initialState), { type, payload }) => {
  switch(type) {
    case AUTH_INITIALIZE: {
      return fromJS(initialState);
    }
    case AUTH_SET_SEED: {
      const { seed, hasPinCode } = payload;
      return state.set('seed', seed).set('hasPinCode', hasPinCode);
    }
    case AUTH_SET_TOKEN: {
      const { token } = payload;
      return state.set('token', token).delete('seed').delete('hasPinCode');
    }

    case AUTH_AUTHENTICATE_BEGIN: {
      return state.set('isAuthenticating', true).delete('error');
    }
    case AUTH_AUTHENTICATE_END: {
      return state.delete('isAuthenticating').delete('error');
    }
    case AUTH_AUTHENTICATE_ERROR: {
      const { error } = payload;
      return state.delete('isAuthenticating').set('error', fromJS(error));
    }

    case AUTH_PAIR_BEGIN: {
      return state.set('isPairing', true).delete('error').delete('hasConnection');
    }
    case AUTH_PAIR_END: {
      return state.delete('isPairing').delete('error').set('hasConnection', true);
    }
    case AUTH_PAIR_ERROR: {
      const { error } = payload;
      return state.delete('isPairing').set('error', fromJS(error)).delete('hasConnection');
    }
    default:
      return state;
  }
};
