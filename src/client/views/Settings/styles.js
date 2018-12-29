import {
  StyleSheet
} from 'react-native';

import { constants } from '../../../../app.json';

export const index = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'stretch'
  },
  menuGroup: {
    borderColor: constants.foregroundColor,
    borderStyle: 'solid',
    borderBottomWidth: 1
  },
  menuItem: {
    backgroundColor: constants.boxOnColor,
    borderColor: constants.foregroundColor,
    borderStyle: 'solid',
    borderTopWidth: 1,
    paddingVertical: 12,
    paddingHorizontal: 16
  },
  menuItemText: {
    fontFamily: constants.fontFamily,
    fontSize: constants.subtitleFontSize,
    fontWeight: constants.subtitleFontWeight,
    color: constants.foregroundColor
  }
});

export const auth = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'stretch'
  },
  row: {
    backgroundColor: constants.boxOnColor,
    borderColor: constants.foregroundColor,
    borderStyle: 'solid',
    borderTopWidth: 1,
    borderBottomWidth: 1,
    paddingVertical: 12,
    paddingHorizontal: 16,
    flexDirection: 'row'
  },
  rowLabel: {
    fontFamily: constants.fontFamily,
    fontSize: constants.subtitleFontSize,
    fontWeight: constants.subtitleFontWeight,
    color: constants.foregroundColor,
    marginRight: 13
  },
  textInput: {
    flex: 1,
    fontFamily: constants.fontFamily,
    fontSize: constants.subtitleFontSize,
    color: constants.textColor,
  },
  codeBox: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'stretch',
    marginTop: 40
  },
  codeDigit: {
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
  codeDigitText: {
    fontWeight: '900',
    fontSize: 40,
    textAlign: 'center',
    fontFamily: constants.fontFamily,
    color: constants.textColor
  }
});
