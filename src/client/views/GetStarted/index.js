import React from 'react';

import {
  Animated,
  StyleSheet,
  Text,
  View
} from 'react-native';
import PropTypes from 'prop-types';
import { connect } from 'react-redux';

import Button from '../../components/Button';
import Logo from '../../components/Logo';
import { constants } from '../../../../app.json';
import ble from '../../services/ble';
import * as actions from '../../actions';
import * as selectors from '../../selectors';

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
    padding: 30
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
  }
});

class GetStarted extends React.Component {
  static contextTypes = {
    logoContainer: PropTypes.object.isRequired
  };

  constructor() {
    super();
    this.progress = new Animated.Value(0);
    this.progress.addListener(this.onProgress);

    this.state = {
      progress: 0
    };
  }

  componentDidMount() {
    const { initialize } = this.props;
    initialize();
  }

  componentWillReceiveProps(nextProps) {
    this.updateProgress(nextProps);
  }

  componentWillUnmount() {
    console.info('unmounting');
  }

  updateProgress(props) {
    const { hasSeed, hasConnection } = props;

    if (this.state.progress < 2 && hasSeed && hasConnection) {
      Animated.timing(this.progress, { toValue: 2 }).start(() => {
        setTimeout(() => {
          Animated.timing(this.progress, { toValue: 3 }).start(() => {
            const { history } = this.props;
            history.replace('/pin-entry');
          });
          this.setState({ progress: 3 });
        }, 3000);
      });
      this.setState({ progress: 2 });
    } else {
      this.onProgress(0, true);

      const { logoContainer } = this.context;
      Animated.timing(logoContainer.houseOpacity, { toValue: 0 }).start();
    }
  }

  onProgress = (value, animated = false) => {
    const { logoContainer } = this.context;
    if (!animated) {
      logoContainer.translateY.setValue(0);
      logoContainer.tagLineOpacity.setValue(1);
    } else {
      Animated.timing(logoContainer.translateY, { toValue: 0 }).start();
      Animated.timing(logoContainer.tagLineOpacity, { toValue: 1 }).start();
    }
  };

  onPressPair = () => {
    const { pair } = this.props;

    Animated.timing(this.progress, { toValue: 1 }).start();
    this.setState({ progress: 1 });

    pair();
  };

  render() {
    const { progress } = this.state;

    const firstPageOpacity = this.progress.interpolate({
      inputRange: [0, 0.5],
      outputRange: [1, 0],
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp'
    });

    const secondPageOpacity = this.progress.interpolate({
      inputRange: [0.5, 1, 1, 2],
      outputRange: [0, 1, 1, 0],
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp'
    });

    const thirdPageOpacity = this.progress.interpolate({
      inputRange: [1, 2, 2.5, 3],
      outputRange: [0, 1, 1, 0],
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp'
    });

    const textOpacity = this.progress.interpolate({
      inputRange: [0, 1, 1.5],
      outputRange: [1, 1, 0],
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp'
    });

    return (
      <View style={styles.container}>
        <View style={{flex: 1}}/>
        <View style={styles.pagesContainer}>
          <View style={styles.blurbContainer}>
            <View style={styles.button}>
              <Animated.View pointerEvents={progress === 0 ? "auto" : "none"} style={[StyleSheet.absoluteFill, { opacity: firstPageOpacity }]}>
                <Button onPress={this.onPressPair} style={styles.button}>PAIR YOUR SMART HUB</Button>
              </Animated.View>
              <Animated.View pointerEvents="none" style={[StyleSheet.absoluteFill, { opacity: secondPageOpacity }]}>
                <Button disabled style={styles.button}>SEARCHING...</Button>
              </Animated.View>
              <Animated.View pointerEvents="none" style={[StyleSheet.absoluteFill, { opacity: thirdPageOpacity, pointerEvents: 'none' }]}>
                <Button disabled style={styles.button}>SUCCESSFULLY PAIRED!</Button>
              </Animated.View>
            </View>
            <Animated.View style={{opacity: textOpacity}}>
              <Text style={styles.blurb}>
                This app requires your <Logo small /> smart hub to be plugged in to a USB charger.
                Please wait for the green light to show and hold down the pairing button until
                the green light starts blinking.
              </Text>
            </Animated.View>
          </View>
        </View>
      </View>
    );
  }
}

const mapDispatchToProps = (dispatch, props) => ({
  initialize: () => dispatch(actions.auth.initialize()),
  pair: () => dispatch(actions.auth.pair()),
  ...props
});
const mapStateToProps = (state, props) => ({
  isPairing: selectors.auth.isPairing(state),
  hasConnection: selectors.auth.hasConnection(state),
  error: selectors.auth.error(state),
  hasSeed: selectors.auth.hasSeed(state),
  ...props
});
export default connect(mapStateToProps, mapDispatchToProps)(GetStarted);
