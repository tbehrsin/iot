import React from 'react';
import { Button, StyleSheet, Text, View } from 'react-native';
import { connect } from 'react-redux';
import { Router, Route, Redirect } from 'react-router-native';
import Stack from 'react-router-native-stack';
import { SharedElementRenderer } from 'react-native-motion';

import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';

import { history } from './router';
import { store, persistor, initialize as initializeStore } from './store';
import { initialize as initializeServices } from './services';
import * as selectors from './selectors';
import * as actions from './actions';
import Logo from './components/Logo';
import Welcome from './views/Welcome';
import GetStarted from './views/GetStarted';
import PinEntry from './views/PinEntry';
import Home from './views/Home';

const PrivateRoute = ({ hasToken, component: Component, ...rest }) => (
  <Route {...rest} render={props => (
    hasToken ? (
      <Component {...props}/>
    ) : (
      <Redirect to={{
        pathname: '/welcome',
        state: { animated: false, next: props.location }
      }}/>
    )
  )}/>
);

let App = ({ location, hasToken }) => (
  <View style={StyleSheet.absoluteFill}>
    <Logo.Container location={location}>
      <Route exact path="/welcome" component={Welcome} />
      <Route exact path="/get-started" component={GetStarted} />
      <Route exact path="/pin-entry" component={PinEntry} />
      <PrivateRoute path="/" component={Home} hasToken={hasToken} />
    </Logo.Container>
  </View>
);

const mapDispatchToProps = (dispatch, props) => ({
  ...props
});
const mapStateToProps = (state, props) => ({
  hasToken: selectors.auth.hasToken(state),
  ...props
});
App = connect(mapStateToProps, mapDispatchToProps)(App);

export default class extends React.Component {
  constructor(initialProps) {
    super(initialProps);
    initializeServices(initialProps);
    initializeStore();
  }

  render() {
    return (
      <Provider store={store}>
        <PersistGate loading={null} persistor={persistor}>
          <Router history={history}>
            <Route component={App} />
          </Router>
        </PersistGate>
      </Provider>
    );
  }
};
