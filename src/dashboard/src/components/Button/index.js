
import React from 'react';
import styles from './index.scss';

class Button extends React.Component {
  render() {
    const { className, component: Component, onClick, children, disabled, ...props } = this.props;

    if (Component) {
      return (
        <Component
          className={`${styles.container} ${className}`}
          onClick={onClick}
          disabled={disabled}
          {...props}
        >
          {children}
        </Component>
      );
    }

    return (
      <button
        className={`${styles.container} ${className}`}
        onClick={onClick}
        disabled={disabled}>
        {children}
      </button>
    )
  }
}

export default Button;
