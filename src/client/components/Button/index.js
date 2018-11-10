
import React from 'react';
import {
  StyleSheet,
  Text,
  TouchableHighlight,
  View
} from 'react-native';
import { constants } from '../../../../app.json';

const styles = StyleSheet.create({
  container: {
    backgroundColor: constants.foregroundColor,
    alignItems: 'center',
    justifyContent: 'center'
  },
  disabled: {
    backgroundColor: null
  },
  text: {
    textAlign: 'center',
    fontFamily: constants.fontFamily,
    fontSize: 21,
    fontWeight: '900',
    color: 'white'
  },
  disabledText: {
    color: constants.foregroundColor
  }
})

class Button extends React.Component {
  render() {
    const { disabled, style, underlayColor = constants.highlightColor, textStyle, onPress, onPressIn, onPressOut } = this.props;

    if (disabled) {
      return (
        <View style={[styles.container, styles.disabled, style]} >
          <Text style={[styles.text, styles.disabledText]}>
            {this.props.children}
          </Text>
        </View>
      );
    }

    return (
      <TouchableHighlight underlayColor={underlayColor} style={[styles.container, style]} onPress={onPress} onPressIn={onPressIn} onPressOut={onPressOut}>
        <View>
          <Text style={[styles.text, textStyle]}>
            {this.props.children}
          </Text>
        </View>
      </TouchableHighlight>
    );
  }
}

export default Button;
