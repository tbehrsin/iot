
import { fromJS } from 'immutable';

const initialState = {
  user: null,
  loggedIn: false
};

export default (state = fromJS(initialState), { type, payload }) => {
  switch(type) {
    case 'auth/SET_USER': {
      const { user } = payload;
      return state.set('user', fromJS(user)).set('loggedIn', !!user);
    }
    default:
      return state;
  }
};
