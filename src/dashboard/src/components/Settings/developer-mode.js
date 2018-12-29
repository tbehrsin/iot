
import React from 'react';
import { connect } from 'react-redux';

import * as actions from '../../actions';
import * as selectors from '../../selectors';

import styles from './developer-mode.scss';

class DeveloperModeSettings extends React.Component {
  componentDidMount() {
    const { request } = this.props;

    request('developer-mode', { path: '/api/v1/developer/' });
    this.createAuthCode();
  }

  componentWillUnmount() {
    clearTimeout(this.timeout);
  }

  createAuthCode = () => {
    const { request } = this.props;
    request('create-auth-code', { path: '/api/v1/auth/code', method: 'POST' });
    this.timeout = setTimeout(this.onCreateAuthCode, 30000);
  };

  onChange = (event) => {
    const { request } = this.props;
    request('developer-mode', { path: '/api/v1/developer/', method: 'POST', body: { enabled: event.target.checked } });
  }

  render() {
    const { developerMode, code } = this.props;
    return (
      <div className={styles.container}>
        <label htmlFor="developerMode">
          <input type="checkbox" name="developerMode" checked={!!developerMode} onChange={this.onChange} />
          <span>Enable developer mode</span>
        </label>
        {developerMode && (
          <div>{code}</div>
        )}
      </div>
    );
  }
}

const mapDispatchToProps = (dispatch, props) => ({
  request: (key, options) => dispatch(actions.api.request(key, options)),
  setIn: (key, path, value) => dispatch(actions.api.setIn(key, path, value)),
  ...props
});
const mapStateToProps = (state, props) => ({
  developerMode: selectors.api.bodyIn(state)('developer-mode', ['enabled']),
  code: selectors.api.bodyIn(state)('create-auth-code', ['code']),
  ...props
});
export default connect(mapStateToProps, mapDispatchToProps)(DeveloperModeSettings);
