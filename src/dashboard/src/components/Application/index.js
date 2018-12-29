
import React from 'react';
import { connect } from 'react-redux';
import { Router, Route, Redirect, Switch, Link } from 'react-router-dom';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { store, persistor } from '../../store';
import { history } from '../../router';

import Routes from '../Page/routes';

class Application extends React.Component {
  render() {
    return (
      <Provider store={store}>
        <PersistGate loading={null} persistor={persistor}>
          <Router history={history}>
            <Routes />
          </Router>
        </PersistGate>
      </Provider>
    );
  }
};

export default Application;
