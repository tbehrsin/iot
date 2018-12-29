
import { createStore, combineReducers, compose, applyMiddleware } from 'redux';
import { persistStore, persistReducer } from 'redux-persist';
import storage from 'redux-persist/lib/storage';
import immutableTransform from 'redux-persist-transform-immutable';
import thunk from 'redux-thunk';
import createSagaMiddleware from 'redux-saga';

import reducers from './reducers';

const persistConfig = {
  transforms: [immutableTransform()],
  key: 'redux',
  storage
};

export const initialize = () => {
  const middleware = [];
  middleware.push(thunk);

  const sagaMiddleware = createSagaMiddleware()
  middleware.push(sagaMiddleware);

  const enhancers = window.__REDUX_DEVTOOLS_EXTENSION__ ? compose(applyMiddleware(...middleware), window.__REDUX_DEVTOOLS_EXTENSION__()) : applyMiddleware(...middleware);

  const reducer = combineReducers(reducers);
  const persistedReducer = persistReducer(persistConfig, reducer);
  const store = createStore(persistedReducer, enhancers);
  const persistor = persistStore(store);

  require('./sagas').default.map(sagaMiddleware.run);
  
  return { store, persistor };
};
