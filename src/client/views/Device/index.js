import React from 'react';

import {
  Animated,
  Image,
  StyleSheet,
  Text,
  TouchableWithoutFeedback,
  View,
  ScrollView,
  Dimensions,
  WebView
} from 'react-native';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import { Route, Switch, Link } from 'react-router-native';
import { constants as config } from '../../../../app.json';
import { ws } from '../../services';
import * as actions from '../../actions';
import * as selectors from '../../selectors';
import { index as styles } from './styles';
import api from './api';

class Device extends React.Component {

  componentDidMount() {
    const { params } = this.props.match;
    const { request } = this.props;

    this.stateSubscription = ws.subscribe(`iot/${params.id}/state`, this.onSetState);

    request(`get-device:${params.id}`, { path: `/api/v1/devices/${params.id}/` });
  }

  componentWillUnmount() {
    this.stateSubscription.unsubscribe();
  }

  onShouldStartLoadWithRequest = (request) => {
    const { url } = this.props;

    // make sure all requests are sandboxed into the controller's router
    return request.url.indexOf(url) === 0;
  };

  onSetState = (topic, state) => {
    const { device, setIn, match } = this.props;
    Object.assign(device.state, state);
    setIn(`get-device:${match.params.id}`, ['state'], state);
  };

  onMessage = (event) => {
    const { params } = this.props.match;
    const { request } = this.props;

    const body = JSON.parse(event.nativeEvent.data);

    request(`patch-device:${params.id}`, { url: `/api/v1/devices/${params.id}/`, method: 'PATCH', body });
  };

  onLoad = () => {
    const { device, url } = this.props;
    const { webView } = this.refs;

    webView.injectJavaScript(`
      document.apiPrefix = ${JSON.stringify(url)};
    `);
    webView.injectJavaScript(api);
    webView.postMessage(device);
  };

  render() {
    const { device, url, authToken } = this.props;

    return (
      <View style={styles.container}>
        { device && (
          <WebView
            ref="webView"
            style={styles.webView}
            onLoad={this.onLoad}
            scalesPagesToFit={false}
            dataDetectorTypes="none"
            scrollEnabled={false}
            geolocationEnabled={true}
            startInLoadingState
            renderLoading={() => <View />}
            onMessage={this.onMessage}
            originWhitelist={['*']}
            source={{
              url: `${url}public/`,
              method: 'POST',
              headers: {
                'X-Authorization': `Bearer ${authToken}`
              }
            }}
            onShouldStartLoadWithRequest={this.onShouldStartLoadWithRequest}
          />
        )}
      </View>
    );
  }
}

const mapDispatchToProps = (dispatch, props) => ({
  request: (key, options) => dispatch(actions.api.request(key, options)),
  setIn: (key, path, value) => dispatch(actions.api.setIn(key, path, value)),
  ...props
});
const mapStateToProps = (state, props) => ({
  device: selectors.api.body(state)(`get-device:${props.match.params.id}`),
  url: selectors.api.url(state)(`get-device:${props.match.params.id}`),
  authToken: selectors.auth.getToken(state),
  ...props
});
export default connect(mapStateToProps, mapDispatchToProps)(Device);
