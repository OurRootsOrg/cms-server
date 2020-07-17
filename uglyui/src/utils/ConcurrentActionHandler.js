export class ConcurrentActionHandler {
  constructor() {
    this._callbacks = [];
  }

  /*
   * Run the supplied action once and return a promise while in progress
   */
  async execute(action) {
    // Create a promise through which to return the result
    const promise = new Promise((resolve, reject) => {
      const onSuccess = () => {
        resolve();
      };

      const onError = error => {
        reject(error);
      };

      this._callbacks.push([onSuccess, onError]);
    });

    // Only do the work for the first UI view that calls us
    if (this._callbacks.length === 1) {
      try {
        // Do the work
        await action();

        // On success resolve all promises
        this._callbacks.forEach(c => {
          c[0]();
        });
      } catch (e) {
        // On failure resolve all promises with the same error
        this._callbacks.forEach(c => {
          c[1](e);
        });
      }

      // Reset once complete
      this._callbacks = [];
    }

    // Return the promise
    return promise;
  }
}
