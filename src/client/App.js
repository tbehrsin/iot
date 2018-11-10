import React from 'react';
import { Button, StyleSheet, Text, View } from 'react-native';
import { connect } from 'react-redux';
import { NativeRouter, Route, Redirect } from 'react-router-native';
import Stack from 'react-router-native-stack';
import { SharedElementRenderer } from 'react-native-motion';

import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';

import { store, persistor, history } from './store';
import * as selectors from './selectors';
import * as actions from './actions';
import Logo from './components/Logo';
import Welcome from './views/Welcome';
import GetStarted from './views/GetStarted';
import PinEntry from './views/PinEntry';
import Home from './views/Home';

const PrivateRoute = ({ loggedIn, component: Component, ...rest }) => (
  <Route {...rest} render={props => (
    loggedIn ? (
      <Component {...props}/>
    ) : (
      <Redirect to={{
        pathname: '/welcome',
        state: { animated: false, next: props.location }
      }}/>
    )
  )}/>
);

let App = ({ location, loggedIn }) => (
  <View style={StyleSheet.absoluteFill}>
    <Logo.Container location={location}>
      <Route exact path="/welcome" component={Welcome} />
      <Route exact path="/get-started" component={GetStarted} />
      <Route exact path="/pin-entry" component={PinEntry} />
      <PrivateRoute path="/" component={Home} loggedIn={loggedIn} />
    </Logo.Container>
  </View>
);

const mapDispatchToProps = (dispatch, props) => ({
  ...props
});
const mapStateToProps = (state, props) => ({
  loggedIn: selectors.selectAuthLoggedIn(state),
  ...props
});
App = connect(mapStateToProps, mapDispatchToProps)(App);

export default () => (
  <Provider store={store}>
    <PersistGate loading={null} persistor={persistor}>
      <NativeRouter>
        <Route component={App} />
      </NativeRouter>
    </PersistGate>
  </Provider>
);
