const {
  MONGO_HOST = "mongo",
  MONGO_PORT = "27017",
  MONGO_BOOTSTRAP_ROOT_USERNAME,
  MONGO_BOOTSTRAP_ROOT_PASSWORD,
  MONGO_ADMIN_USERNAME,
  MONGO_ADMIN_PASSWORD,
  MONGO_DEPLOYMENT_MARKER = "aeshield-deployment-managed",
  MONGO_RECONCILE_AUTH_MODE = "bootstrap-root",
} = process.env;

function fail(message) {
  print(message);
  quit(1);
}

function requireEnv(name, value) {
  if (!value) {
    fail(`Missing required environment variable: ${name}`);
  }
}

requireEnv("MONGO_BOOTSTRAP_ROOT_USERNAME", MONGO_BOOTSTRAP_ROOT_USERNAME);
requireEnv("MONGO_BOOTSTRAP_ROOT_PASSWORD", MONGO_BOOTSTRAP_ROOT_PASSWORD);
requireEnv("MONGO_ADMIN_USERNAME", MONGO_ADMIN_USERNAME);
requireEnv("MONGO_ADMIN_PASSWORD", MONGO_ADMIN_PASSWORD);

const encodedRootUser = encodeURIComponent(MONGO_BOOTSTRAP_ROOT_USERNAME);
const encodedRootPassword = encodeURIComponent(MONGO_BOOTSTRAP_ROOT_PASSWORD);
const encodedAdminUser = encodeURIComponent(MONGO_ADMIN_USERNAME);
const encodedAdminPassword = encodeURIComponent(MONGO_ADMIN_PASSWORD);
const rootUri = `mongodb://${encodedRootUser}:${encodedRootPassword}@${MONGO_HOST}:${MONGO_PORT}/admin?authSource=admin`;
const desiredUri = `mongodb://${encodedAdminUser}:${encodedAdminPassword}@${MONGO_HOST}:${MONGO_PORT}/admin?authSource=admin`;
const reconcileUri = MONGO_RECONCILE_AUTH_MODE === "deploy-admin" ? desiredUri : rootUri;

function connect(uri) {
  return new Mongo(uri).getDB("admin");
}

function getManagedUser(adminDb) {
  const usersInfo = adminDb.runCommand({ usersInfo: 1, showCustomData: true });
  if (!usersInfo.ok) {
    fail(`Unable to load MongoDB users: ${usersInfo.errmsg || "unknown error"}`);
  }

  const matches = (usersInfo.users || []).filter(
    (user) => user.customData?.managedBy === MONGO_DEPLOYMENT_MARKER,
  );

  if (matches.length > 1) {
    fail("Found multiple deployment-managed MongoDB admin users; aborting reconciliation.");
  }

  return matches[0] || null;
}

function desiredCredentialsWork() {
  try {
    const db = connect(desiredUri);
    const result = db.runCommand({ connectionStatus: 1 });
    return Boolean(result.ok);
  } catch (error) {
    return false;
  }
}

const adminDb = connect(reconcileUri);
const managedUser = getManagedUser(adminDb);
const desiredExistsInfo = adminDb.runCommand({ usersInfo: MONGO_ADMIN_USERNAME, showCustomData: true });
if (!desiredExistsInfo.ok) {
  fail(`Unable to check desired MongoDB admin user: ${desiredExistsInfo.errmsg || "unknown error"}`);
}
const desiredExistingUser = desiredExistsInfo.users?.[0] || null;

if (!managedUser) {
  if (desiredExistingUser) {
    if (desiredCredentialsWork()) {
      print(`MongoDB admin user '${MONGO_ADMIN_USERNAME}' already matches desired credentials.`);
      quit(0);
    }

    adminDb.updateUser(MONGO_ADMIN_USERNAME, {
      pwd: MONGO_ADMIN_PASSWORD,
      roles: desiredExistingUser.roles,
      customData: desiredExistingUser.customData || undefined,
    });
    print(`MongoDB admin user '${MONGO_ADMIN_USERNAME}' password updated.`);
    quit(0);
  }

  adminDb.createUser({
    user: MONGO_ADMIN_USERNAME,
    pwd: MONGO_ADMIN_PASSWORD,
    roles: [{ role: "root", db: "admin" }],
    customData: { managedBy: MONGO_DEPLOYMENT_MARKER },
  });
  print(`MongoDB admin user '${MONGO_ADMIN_USERNAME}' created.`);
  quit(0);
}

if (managedUser.user === MONGO_ADMIN_USERNAME) {
  if (desiredCredentialsWork()) {
    print(`MongoDB admin user '${MONGO_ADMIN_USERNAME}' already matches desired credentials.`);
    quit(0);
  }

  adminDb.updateUser(MONGO_ADMIN_USERNAME, {
    pwd: MONGO_ADMIN_PASSWORD,
    roles: managedUser.roles,
    customData: { ...(managedUser.customData || {}), managedBy: MONGO_DEPLOYMENT_MARKER },
  });
  print(`MongoDB admin user '${MONGO_ADMIN_USERNAME}' password updated.`);
  quit(0);
}

adminDb.createUser({
  user: MONGO_ADMIN_USERNAME,
  pwd: MONGO_ADMIN_PASSWORD,
  roles: [{ role: "root", db: "admin" }],
  customData: { managedBy: MONGO_DEPLOYMENT_MARKER },
});

if (!desiredCredentialsWork()) {
  fail(`MongoDB replacement user '${MONGO_ADMIN_USERNAME}' failed credential verification after creation.`);
}

const removed = adminDb.dropUser(managedUser.user);
if (!removed) {
  fail(`Failed to remove previous deployment-managed MongoDB admin user '${managedUser.user}'.`);
}

print(`MongoDB admin user '${managedUser.user}' replaced with '${MONGO_ADMIN_USERNAME}'.`);
quit(0);
