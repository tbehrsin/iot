
import React from 'react';
import { Switch, Route, Redirect } from 'react-router-dom';
import { TransitionGroup, CSSTransition } from 'react-transition-group';

import { connect } from 'react-redux';

import * as actions from '../../actions';
import * as selectors from '../../selectors';

import Home from '../Home';
import Zone from '../Zone';
import Device from '../Device';
import Assign from '../Assign';
import Settings from '../Settings';
import Connect from '../Connect';


const PrivateRoute = ({ authenticated, component: Component, ...rest }) => (
  <Route {...rest} render={props => (
    authenticated ? (
      <Component {...props}/>
    ) : (
      <Redirect to={{
        pathname: '/connect',
        state: { animated: false, next: props.location }
      }}/>
    )
  )}/>
);

const mapDispatchToProps = (dispatch, props) => ({
  setToken: token => dispatch(actions.auth.setToken(token)),
  ...props
});
const mapStateToProps = (state, props) => ({
  hasAppToken: selectors.auth.hasAppToken(state),
  ...props
});

export default connect(mapStateToProps, mapDispatchToProps, null, { pure: false })(({ hasAppToken }) => (
  <Switch>
    <Route exact path="/connect" component={Connect} />
    <Redirect exact path="/dashboard" to="/dashboard/devices/" />
    <PrivateRoute exact path="/dashboard/devices/" component={Zone} authenticated={hasAppToken} />
    <PrivateRoute exact path="/dashboard/devices/:id" component={Device} authenticated={hasAppToken} />
    <PrivateRoute exact path="/dashboard/devices/:id/assign" component={Assign} authenticated={hasAppToken} />
    <PrivateRoute path="/dashboard/settings/" component={Settings} authenticated={hasAppToken} />
    <Route exact path="/" component={Home} />
    <Redirect path="/dashboard/" to="/dashboard" />
    <Redirect path="/" to="/" />
  </Switch>
));
