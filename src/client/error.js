
// https://stackoverflow.com/questions/31089801/extending-error-in-javascript-with-es6-syntax-babel
class ResourceError extends Error {
  constructor({ message, code }) {
    super();
    this.name = this.constructor.name;
    this.code = code;
    this.message = message;
    if (typeof Error.captureStackTrace === 'function') {
      Error.captureStackTrace(this, this.constructor);
    } else {
      this.stack = (new Error()).stack;
    }
  }

  toJSON() {
    return {
      message: this.message,
      code: this.code
    };
  }
}

ResourceError.BadRequest = 400;
ResourceError.Unauthorized = 401;
ResourceError.Forbidden = 403;
ResourceError.NotFound = 404;
ResourceError.MethodNotAllowed = 405;
ResourceError.InternalServerError = 500;

global.ResourceError = ResourceError;
