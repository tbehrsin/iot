

export const authSetUser = (user) => (dispatch) => (
  dispatch({
    type: 'auth/SET_USER',
    payload: {
      user
    }
  })
);
