## Lambda Function to KMS Key

Grants a Lambda function cryptographic use of a KMS key.

When a function links to a key, the function's execution role is granted use of the key and (by default) an environment variable holding the key's ARN is added to the function so your code can reference it at runtime without manual configuration or hardcoding.

The access level is controlled with the `aws.lambda.kms.<targetKey>.accessLevel` annotation:

- `decrypt` (default) — `kms:Decrypt`, `kms:DescribeKey`
- `encryptDecrypt` — adds `kms:Encrypt`, `kms:GenerateDataKey`, `kms:GenerateDataKeyWithoutPlaintext`

Each statement grants access to the specific key ARN.

The function's execution role must be defined as a resource in the same blueprint; the link adds the access permission to it.

KMS authorises requests against **both** the key policy and the caller's IAM policy. This link grants the IAM (identity) side and relies on the key policy delegating authorisation to IAM, the default statement AWS adds to newly created keys ("Enable IAM policies", allowing `kms:*` to the account root). If the key policy has been locked down to remove that delegation, this link's IAM grant alone is insufficient.

For that case, set `aws.lambda.kms.<targetKey>.manageKeyGrant` to `true`. The link then also creates a **KMS grant** for the function's execution role covering the `accessLevel` operations, which authorises use of the key independently of the key policy's IAM delegation. The grant is reconciled by a deterministic name unique to the function↔key pair, and is revoked when the link is destroyed or the annotation is set back to `false`. This uses KMS grants rather than editing the key policy, so the key's declarative policy is never modified. Grant creation, update and revocation are surfaced in the deployment's staged changes (as a `keyGrant` link field) so the side effect is visible before applying. Defaults to `false`; when left off, if the key defines a custom policy that does not appear to delegate to IAM, the link surfaces a validation warning at plan time.

If the function runs inside a VPC without internet access, the link also activates a KMS interface VPC endpoint on the linked flex VPC so the function can reach KMS over the private network.

### Example

```blueprintlang
version "2025-11-02"

resource encryptFunction: aws/lambda/function {
    metadata {
        labels = {
            key = "data-encryption"
        }
        annotations = {
            "aws.lambda.kms.dataKey.envVarName" = "DATA_ENCRYPTION_KEY"
            "aws.lambda.kms.dataKey.accessLevel" = "encryptDecrypt"
        }
    }

    select by label {
        key = "data-encryption"
    }

    spec {
        functionName = "encrypt"
        role = encryptFunctionRole.spec.arn
        # ... other function configuration
    }
}

resource dataKey: aws/kms/key {
    metadata {
        labels = {
            key = "data-encryption"
        }
    }

    spec {
        description = "Data encryption key"
    }
}

resource encryptFunctionRole: aws/iam/role {
    spec {
        name = "encrypt-role"
        # ... role configuration
    }
}
```
