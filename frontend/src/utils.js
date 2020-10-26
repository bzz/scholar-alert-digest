const dispatchAction = dispatch => type => payload => dispatch({type, payload})

export const getActions = handlers => dispatch => Object.keys(handlers).reduce(
  (acc, x) => ({...acc, [x]: dispatchAction(dispatch)(x)}), {},
)

export const createReducer = handlers => (state, action) => {
  // eslint-disable-next-line
  if (handlers.hasOwnProperty(action.type)) {
    return handlers[action.type](state, action)
  }

  return state
}
