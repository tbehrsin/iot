
import { fromJS } from 'immutable';
import {
  AUTH_SET_TOKEN,
} from '../constants';

const initialState = {
};

export default (state = fromJS(initialState), { type, payload }) => {
  switch(type) {
    case AUTH_SET_TOKEN: {
      const { token } = payload;
      return state.set('token', token);
    }
    default:
      return state;
  }
};
