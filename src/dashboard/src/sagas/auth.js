

import { takeEvery, put, select } from 'redux-saga/effects';
import * as constants from '../constants';
import * as selectors from '../selectors';
import * as actions from '../actions';
import { history } from '../router';

function* convertAuthToken({ payload }) {
  if (payload.key !== 'convert-auth-token') {
    return;
  }

  yield put(actions.auth.setToken(payload.body.token));
  history.go(-history.index);
  history.replace('/');
}

export function* watchConvertAuthToken() {
  yield takeEvery(constants.API_REQUEST_END, convertAuthToken);
}
