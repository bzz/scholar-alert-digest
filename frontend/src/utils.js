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

export const fromMaybe = def => f => {
  try {
    const res = f()
    return res != null ? res : def
  } catch (e) {
    return def
  }
}

export const Maybe = ({cond, children}) => {
  const a = typeof cond === "function"

  if (a) {
    return fromMaybe(false)(cond) ? fromMaybe(children)(children) : null
  }

  return cond ? fromMaybe(children)(children) : null
}

export const Either = ({cond, children}) => (
  Maybe({cond, children: children[0]}) || fromMaybe(children[1])(children[1])
)
