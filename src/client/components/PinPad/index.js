
import React from 'react';
import {
  Animated,
  Image,
  View,
  StyleSheet,
  Dimensions,
  Text
} from 'react-native';
import Button from '../Button';
import { constants } from '../../../../app.json';

// https://fontawesome.com/license
import BackspaceIcon from './backspace.png';

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    alignItems: 'stretch',
    borderTopColor: constants.textColor,
    borderTopWidth: StyleSheet.hairlineWidth,
    borderStyle: 'solid',
  },
  row: {
    flexDirection: 'row',
    alignItems: 'stretch'
  },
  button: {
    flex: 1,
    backgroundColor: '#eee',
    padding: 20
  },
  buttonText: {
    color: constants.textColor
  }
});

const dim = Dimensions.get('window');

class PinPad extends React.Component {

  constructor() {
    super();
    this.state = {
      translateY: null
    };
    this.show = new Animated.Value(0);
  }

  componentWillReceiveProps(nextProps) {
    console.info(nextProps);
    Animated.timing(this.show, { toValue: nextProps.show ? 1 : 0 }).start();
  }

  onLayout = ({ nativeEvent: event }) => {
    const { layout: { height } } = event;

    const translateY = this.show.interpolate({
      inputRange: [0, 1],
      outputRange: [height, 0],
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp'
    })

    this.setState({ translateY });
  }

  render() {
    const { onPress = () => {}, onPressBackspace = () => {}, show } = this.props;
    const { translateY } = this.state;

    const underlayColor = "#ddd";

    return (
      <Animated.View onLayout={this.onLayout} style={[styles.container, { transform: [{translateY: translateY || dim.height }]}]}>
        <View style={styles.row}>
          <Button onPressIn={() => onPress(1)} underlayColor={underlayColor} style={styles.button} textStyle={styles.buttonText}>1</Button>
          <Button onPressIn={() => onPress(2)} underlayColor={underlayColor} style={styles.button} textStyle={styles.buttonText}>2</Button>
          <Button onPressIn={() => onPress(3)} underlayColor={underlayColor} style={styles.button} textStyle={styles.buttonText}>3</Button>
        </View>
        <View style={styles.row}>
          <Button onPressIn={() => onPress(4)} underlayColor={underlayColor} style={styles.button} textStyle={styles.buttonText}>4</Button>
          <Button onPressIn={() => onPress(5)} underlayColor={underlayColor} style={styles.button} textStyle={styles.buttonText}>5</Button>
          <Button onPressIn={() => onPress(6)} underlayColor={underlayColor} style={styles.button} textStyle={styles.buttonText}>6</Button>
        </View>
        <View style={styles.row}>
          <Button onPressIn={() => onPress(7)} underlayColor={underlayColor} style={styles.button} textStyle={styles.buttonText}>7</Button>
          <Button onPressIn={() => onPress(8)} underlayColor={underlayColor} style={styles.button} textStyle={styles.buttonText}>8</Button>
          <Button onPressIn={() => onPress(9)} underlayColor={underlayColor} style={styles.button} textStyle={styles.buttonText}>9</Button>
        </View>
        <View style={styles.row}>
          <View style={styles.button}/>
          <Button onPressIn={() => onPress(0)} underlayColor={underlayColor} style={styles.button} textStyle={styles.buttonText}>0</Button>
          <Button onPressIn={() => onPressBackspace()} underlayColor={underlayColor} style={styles.button} textStyle={[styles.buttonText, {marginBottom: -8}]}>
            <Image source={BackspaceIcon} />
          </Button>
        </View>
        </Animated.View>
    );
  }
}

export default PinPad;
