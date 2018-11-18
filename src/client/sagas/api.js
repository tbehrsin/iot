
import { takeEvery, put, select } from 'redux-saga/effects';
import { AUTH_AUTHENTICATE_END } from '../constants';
import * as selectors from '../selectors';
import * as actions from '../actions';

import { ws } from '../services';

const PERSIST_REHYDRATE = 'persist/REHYDRATE';

function* connect(action) {
  const token = yield select(selectors.auth.getToken);

  if (token) {
    yield put(actions.api.connect());
    ws.subscribe('#', (topic, payload) => {
      console.info(topic, payload);
    });
    ws.subscribe('gw/#', (topic, payload) => {
      console.info(topic, payload);
    });

    ws.subscribe('gw/90FD9FFFFE7B59D0/heartbeat', (topic, payload) => {
      console.info(topic, payload);
    });

    ws.publish('gw/90FD9FFFFE7B59D0/publishstate');
  } else {
    yield put(actions.api.disconnect());
  }
}

export function* watchAuthEnd() {
  yield takeEvery(AUTH_AUTHENTICATE_END, connect);
}

export function* watchPersistRehydrate() {
  yield takeEvery(PERSIST_REHYDRATE, connect);
}
