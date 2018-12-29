import React from 'react';

import {
  StyleSheet,
  Text,
  TouchableWithoutFeedback,
  View,
  TextInput
} from 'react-native';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import { constants } from '../../../../app.json';
import * as actions from '../../actions';
import * as selectors from '../../selectors';
import { auth as styles } from './styles';

class Auth extends React.Component {
  constructor() {
    super();
    this.state = {
      email: ''
    };
  }

  componentDidMount() {
    this.onCreateAuthCode();
  }

  componentWillUnmount() {
    clearTimeout(this.timeout);
  }

  onCreateAuthCode = () => {
    const { request } = this.props;
    request('create-auth-code', { path: '/api/v1/auth/code', method: 'POST' });
    setTimeout(this.onCreateAuthCode, 30000);
  };

  onChangeEmail = (email) => {
    this.setState({ email });
  };

  onSubmit = (event) => {
    const { request } = this.props;
    const { email } = this.state;

    console.info(email);
    request(`create-email-token`, { method: 'POST', path: '/api/v1/auth/', body: { email } });
  };

  render() {
    const { device, url, authToken, authCode } = this.props;
    const { email } = this.state;

    return (
      <View style={styles.container}>
        <View style={styles.row}>
          <Text style={styles.rowLabel}>Email</Text>
          <TextInput autoCapitalize="none" autoCorrect={false} autoFocus clearTextOnFocus={true} keyboardAppearance="light" keyboardType="email-address" value={email} onChangeText={this.onChangeEmail} onSubmitEditing={this.onSubmit} returnKeyType="done" selectionColor={constants.foregroundColor} spellCheck={false} textContentType="emailAddress" underlineColorAndroid="transparent" style={styles.textInput}/>
        </View>

        {authCode && <View style={styles.codeBox}>
          <View style={[styles.codeDigit, {marginLeft: 0}]}><Text style={styles.codeDigitText}>{authCode.code[0]}</Text></View>
          <View style={styles.codeDigit}><Text style={styles.codeDigitText}>{authCode.code[1]}</Text></View>
          <View style={styles.codeDigit}><Text style={styles.codeDigitText}>{authCode.code[2]}</Text></View>
          <View style={styles.codeDigit}><Text style={styles.codeDigitText}>{authCode.code[3]}</Text></View>
          <View style={styles.codeDigit}><Text style={styles.codeDigitText}>{authCode.code[4]}</Text></View>
          <View style={styles.codeDigit}><Text style={styles.codeDigitText}>{authCode.code[5]}</Text></View>
        </View>}
      </View>
    );
  }
}

const mapDispatchToProps = (dispatch, props) => ({
  request: (key, options) => dispatch(actions.api.request(key, options)),
  ...props
});
const mapStateToProps = (state, props) => ({
  authCode: selectors.api.body(state)('create-auth-code'),
  ...props
});
export default connect(mapStateToProps, mapDispatchToProps)(Auth);
