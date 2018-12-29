
import { takeEvery, put, select } from 'redux-saga/effects';
import { delay } from "redux-saga";
import { AUTH_SET_TOKEN, API_REQUEST_ERROR, API_CONNECT_ERROR, API_CONNECT_END } from '../constants';

import { constants } from '../../app.json';
import * as selectors from '../selectors';
import * as actions from '../actions';
import { ws } from '../services';

const PERSIST_REHYDRATE = 'persist/REHYDRATE';

function* connect() {
  const hasAppToken = yield select(selectors.auth.hasAppToken);

  if (hasAppToken) {
    yield put(actions.api.connect());
  } else {
    yield put(actions.api.disconnect());
  }
}

export function* watchSetToken() {
  yield takeEvery(AUTH_SET_TOKEN, connect);
}

export function* watchPersistRehydrate() {
  yield takeEvery(PERSIST_REHYDRATE, connect);
}

function* refreshRequests() {
  const requests = yield select(selectors.api.refreshRequests);

  for (const [key, options] of Object.entries(requests)) {
    yield put(actions.api.request(key, options));
  }
}

export function* watchConnectEnd() {
  yield takeEvery(API_CONNECT_END, refreshRequests);
}

function* setPublic(action) {
  if (!constants.forceServer && (yield select(selectors.api.local))) {
    yield put(actions.api.setLocal(false));
    yield delay(5000);
    yield put(actions.api.setLocal(true));
  }
}

export function* watchRequestError() {
  yield takeEvery(API_REQUEST_ERROR, setPublic);
}

export function* watchConnectError() {
  yield takeEvery(API_CONNECT_ERROR, setPublic);
}
