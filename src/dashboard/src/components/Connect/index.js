
import React from 'react';
import { connect } from 'react-redux';
import { Router, Route, Redirect } from 'react-router-dom';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { store, persistor } from '../../store';
import { history } from '../../router';
import Page from '../Page';

import * as actions from '../../actions';
import * as selectors from '../../selectors';

import styles from './index.scss';

class Connect extends React.Component {

  constructor() {
    super();
    this.state = {
      code: ''
    };
  }

  componentDidMount() {
    window.addEventListener('hashchange', this.onHashChange, false);
    this.onHashChange();
  }

  componentWillUnmount() {
    window.removeEventListener('hashchange', this.onHashChange, false);
  }

  onHashChange = () => {
    const { setToken, hasEmailToken } = this.props;

    let match;
    if (match = location.hash.match(/^#token=(.*)$/)) {
      setToken(match[1]);
      location.hash = '';
    } else if(!hasEmailToken){
      history.go(-history.index);
      history.replace('/');
    }
  };

  onChange = (i, value) => {
    if (!/[a-z0-9]/ig.test(value)) {
      return
    }

    value = value.replace(/I/i, '1');
    value = value.replace(/O/i, '0');

    let { code } = this.state;
    code = code.substring(0, i) + value.substring(0, 1).toUpperCase();
    this.setState({ code });
    if (code.length < 6) {
      this.focusing = true;
      this.refs.code.querySelectorAll('input')[code.length].focus();
      this.focusing = false;
    } else {
      this.refs.code.querySelectorAll('input')[code.length - 1].blur();

      const { request } = this.props;
      request('convert-auth-token', { method: 'PUT', path: '/api/v1/auth/', body: { code } });
    }
  };

  onFocus = (i) => {
    if (this.focusing) {
      return;
    }

    let { code } = this.state;
    code = code.substring(0, i);
    this.setState({ code });
    this.refs.code.querySelectorAll('input')[code.length].focus();
  };

  onKeyDown = (i, event) => {
    if (event.keyCode === 0x08) {
      this.onFocus(i - 1);
    }
  }

  render() {
    const { code, location } = this.state;

    return (
      <Page.FullScreen className={styles.container} location={location}>
        <p>Enter the code found in the Behrsin IoT app to secure your account:</p>
        <div ref="code" className={styles.code}>
          {Array(6).fill().map((_, i) => (
            <input
              key={i}
              type="text"
              autoFocus={i === 0}
              value={code[i] || ''}
              onChange={(event) => this.onChange(i, event.target.value)}
              onKeyDown={(event) => this.onKeyDown(i, event)}
              onFocus={() => this.onFocus(i)}
            />
          ))}
        </div>
      </Page.FullScreen>
    );
  }
}

const mapDispatchToProps = (dispatch, props) => ({
  request: (key, options) => dispatch(actions.api.request(key, options)),
  setToken: (token) => dispatch(actions.auth.setToken(token)),
  ...props
});
const mapStateToProps = (state, props) => ({
  authToken: selectors.api.body(state)('convert-auth-token'),
  hasEmailToken: selectors.auth.hasEmailToken(state),
  ...props
});
export default connect(mapStateToProps, mapDispatchToProps)(Connect);
