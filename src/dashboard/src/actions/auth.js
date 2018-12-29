
import {
  AUTH_SET_TOKEN,
} from '../constants';

export const setToken = (token) => ({
  type: AUTH_SET_TOKEN,
  payload: { token }
});
