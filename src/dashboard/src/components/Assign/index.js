
import React from 'react';
import { connect } from 'react-redux';
import { Router, Route, Redirect, Link } from 'react-router-dom';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { store, persistor } from '../../store';
import { history } from '../../router';

import * as actions from '../../actions';
import * as selectors from '../../selectors';
import { ws } from '../../services';

import Page from '../Page';
import Button from '../Button';
import Accordion from '../Accordion';
import DeveloperMode from '../DeveloperMode';

import styles from './index.scss';


class Assign extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
    };
  }

  componentDidMount() {
    const { device, request, match } = this.props;

    this.stateSubscription = ws.subscribe(`iot/${match.params.id}/notify`, this.onNotify);

    if (!device) {
      request('devices', { path: '/api/v1/devices/' });
    }
    request('developer-mode', { path: '/api/v1/developer/' });
  }

  componentWillUnmount() {
    this.stateSubscription.unsubscribe();
  }

  onNotify = (topic, device) => {

  };

  render() {
    const { device, url, developerMode } = this.props;

    return (
      <Page.FullScreen className={styles.container}>
        {device && (
          <div className={styles.columns}>
            <div>
              <div className={styles.field}><h3>Manufacturer</h3> {device.manufacturer}</div>
              <div className={styles.field}><h3>Model</h3> {device.model}</div>
            </div>
            <div>
              {developerMode && developerMode.enabled && (
                <DeveloperMode device={device} />
              )}
            </div>
          </div>
        )}
      </Page.FullScreen>
    );
  }
}

const mapDispatchToProps = (dispatch, props) => ({
  request: (key, options) => dispatch(actions.api.request(key, options)),
  setIn: (key, path, value) => dispatch(actions.api.setIn(key, path, value)),
  ...props
});
const mapStateToProps = (state, props) => ({
  device: selectors.api.findInBody(state)('devices', d => d.get('id') === props.match.params.id),
  url: selectors.api.url(state)('devices'),
  developerMode: selectors.api.body(state)('developer-mode'),
  ...props
});
export default connect(mapStateToProps, mapDispatchToProps)(Assign);
