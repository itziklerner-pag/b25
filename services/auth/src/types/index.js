// Type definitions for the authentication service
// These were interfaces in TypeScript, now documented via JSDoc comments

/**
 * @typedef {Object} User
 * @property {string} id
 * @property {string} email
 * @property {string} password_hash
 * @property {Date} created_at
 * @property {Date} updated_at
 * @property {Date} [last_login]
 * @property {boolean} is_active
 */

/**
 * @typedef {Object} UserRegistration
 * @property {string} email
 * @property {string} password
 */

/**
 * @typedef {Object} UserLogin
 * @property {string} email
 * @property {string} password
 */

/**
 * @typedef {Object} TokenPayload
 * @property {string} userId
 * @property {string} email
 * @property {number} [iat]
 * @property {number} [exp]
 */

/**
 * @typedef {Object} AuthTokens
 * @property {string} accessToken
 * @property {string} refreshToken
 * @property {number} expiresIn
 */

/**
 * @typedef {Object} RefreshToken
 * @property {string} id
 * @property {string} user_id
 * @property {string} token_hash
 * @property {Date} expires_at
 * @property {Date} created_at
 * @property {boolean} revoked
 */

/**
 * @typedef {Object} ApiResponse
 * @property {boolean} success
 * @property {*} [data]
 * @property {Object} [error]
 * @property {string} error.code
 * @property {string} error.message
 * @property {*} [error.details]
 * @property {string} timestamp
 */

/**
 * @typedef {Object} DatabaseConfig
 * @property {string} host
 * @property {number} port
 * @property {string} database
 * @property {string} user
 * @property {string} password
 * @property {number} max
 * @property {number} idleTimeoutMillis
 * @property {number} connectionTimeoutMillis
 */

/**
 * @typedef {Object} JWTConfig
 * @property {string} accessTokenSecret
 * @property {string} refreshTokenSecret
 * @property {string} accessTokenExpiry
 * @property {string} refreshTokenExpiry
 */

/**
 * @typedef {Object} ServerConfig
 * @property {number} port
 * @property {string} nodeEnv
 * @property {string[]} corsOrigins
 * @property {number} rateLimitWindowMs
 * @property {number} rateLimitMaxRequests
 */

// Export empty object (types are only for documentation)
export {};
