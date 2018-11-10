import React from 'react';

import {
  Animated,
  StyleSheet,
  Text,
  TouchableWithoutFeedback,
  View
} from 'react-native';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import { SharedElement } from 'react-native-motion';
import * as actions from '../../actions';
import * as selectors from '../../selectors';
import Button from '../../components/Button';
import Logo from '../../components/Logo';
import PinPad from '../../components/PinPad';
import { constants } from '../../../../app.json';

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'stretch'
  },
  logoContainer: {
    flex: 1,
    justifyContent: 'flex-end',
    alignItems: 'center',
  },
  pagesContainer: {
    flex: 1,
    justifyContent: 'flex-start',
    alignItems: 'stretch',
    paddingBottom: 90
  },
  blurbContainer: {
    justifyContent: 'flex-start',
    alignItems: 'stretch',
    padding: 30,
    paddingBottom: 60
  },
  blurb: {
    fontFamily: constants.fontFamily,
    fontSize: 20,
    lineHeight: 32,
    alignSelf: 'center',
    color: constants.textColor,
    textAlign: 'center',
    maxWidth: 315,
    marginTop: 30
  },
  button: {
    height: 62
  },
  pinInputBox: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'stretch',
    marginTop: 40
  },
  pinDigit: {
    borderStyle: 'solid',
    borderRadius: 4,
    borderColor: constants.textColor,
    borderWidth: 1,
    backgroundColor: 'white',
    alignItems: 'center',
    justifyContent: 'center',
    marginLeft: 8,
    width: 47,
    height: 70
  },
  pinDigitText: {
    fontWeight: '900',
    fontSize: 40,
    textAlign: 'center',
    fontFamily: constants.fontFamily,
    color: constants.textColor
  }
});

class PinEntry extends React.Component {
  static contextTypes = {
    logoContainer: PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.progress = new Animated.Value(3);
    this.progress.addListener(this.onProgress);

    this.state = {
      progress: 3,
      pin: '',
      showPinPad: false
    };
  }

  componentDidMount() {
    const { logoContainer } = this.context;
    console.info(this.props.location);
    this.onProgress(0, true);
    Animated.timing(logoContainer.houseOpacity, { toValue: 0 }).start();
  }

  onProgress = (value, animated = false) => {
    const { logoContainer } = this.context;
    if (!animated) {
      logoContainer.translateY.setValue(-90);
      logoContainer.tagLineOpacity.setValue(0);
    } else {
      Animated.timing(logoContainer.translateY, { toValue: -90, duration: 400 }).start();
      Animated.timing(logoContainer.tagLineOpacity, { toValue: 0, duration: 400 }).start();
    }
  };

  onPressNumber = (value) => {
    let { pin } = this.state;
    pin += value;
    pin = pin.substring(0, 6);
    this.setState({ pin });
    if (pin.length === 6) {
      const { history, authSetUser } = this.props;
      authSetUser({});
      history.replace('/home');
    }
  };

  onPressBackspace = () => {
    let { pin } = this.state;
    pin = pin.substring(0, pin.length - 1);
    this.setState({ pin });
  };

  onPressPinInputBox = () => {
    this.setState({ showPinPad: true });
  };

  render() {
    const { progress, pin, showPinPad } = this.state;

    return (
      <View style={styles.container}>
        <View style={{flex: 1, marginBottom: -380}}/>
        <View style={{flex: 1}}>
          <Button disabled style={styles.button}>SETUP YOUR SMART HUB</Button>
          <TouchableWithoutFeedback onPressIn={this.onPressPinInputBox}>
            <View style={styles.pinInputBox}>
              <View style={[styles.pinDigit, {marginLeft: 0}]}><Text style={styles.pinDigitText}>{pin[0]}</Text></View>
              <View style={styles.pinDigit}><Text style={styles.pinDigitText}>{pin[1]}</Text></View>
              <View style={styles.pinDigit}><Text style={styles.pinDigitText}>{pin[2]}</Text></View>
              <View style={styles.pinDigit}><Text style={styles.pinDigitText}>{pin[3]}</Text></View>
              <View style={styles.pinDigit}><Text style={styles.pinDigitText}>{pin[4]}</Text></View>
              <View style={styles.pinDigit}><Text style={styles.pinDigitText}>{pin[5]}</Text></View>
            </View>
          </TouchableWithoutFeedback>
          <View style={{flex: 1}} />
          <View style={styles.blurbContainer}>
            <Text style={styles.blurb}>
              Enter a PIN code to make sure you can gain access to this device in the future.
              This can be shared with your family so they can pair with your <Logo small /> smart
              hub too. Please keep this PIN code safe.
            </Text>
          </View>
        </View>
        {showPinPad && (
          <View style={StyleSheet.absoluteFill}>
            <TouchableWithoutFeedback onPressIn={() => this.setState({ showPinPad: false })}>
              <View style={StyleSheet.absoluteFill} />
            </TouchableWithoutFeedback>
          </View>
        )}
        <PinPad onPress={this.onPressNumber} onPressBackspace={this.onPressBackspace} show={showPinPad} />
      </View>
    );
  }
}

const mapDispatchToProps = (dispatch, props) => ({
  authSetUser: user => dispatch(actions.authSetUser(user)),
  ...props
});
const mapStateToProps = (state, props) => ({
  ...props
});
export default connect(mapStateToProps, mapDispatchToProps)(PinEntry);
