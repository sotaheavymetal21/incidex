"use client";

import { useState } from "react";
import { useAuth } from "@/context/AuthContext";
import { userApi } from "@/lib/api";
import { User, Role, UpdateUserRequest } from "@/types/user";
import {
  ValidationLimits,
  containsEmoji,
  isValidEmployeeNumber,
  isValidEmail,
} from "@/utils/validation";

interface EditUserModalProps {
  user: User;
  onClose: () => void;
  onSuccess: () => void;
}

interface FieldErrors {
  name?: string;
  email?: string;
  employee_number?: string;
  department?: string;
}

export default function EditUserModal({
  user,
  onClose,
  onSuccess,
}: EditUserModalProps) {
  const { token } = useAuth();
  const [formData, setFormData] = useState<UpdateUserRequest>({
    name: user.name,
    email: user.email,
    employee_number: user.employee_number,
    department: user.department,
    role: user.role,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});

  const validateField = (
    field: keyof FieldErrors,
    value: string,
  ): string | undefined => {
    switch (field) {
      case "name":
        if (!value.trim()) return "名前は必須です";
        if (value.length > ValidationLimits.NAME_MAX_LENGTH)
          return `名前は${ValidationLimits.NAME_MAX_LENGTH}文字以内で入力してください`;
        if (containsEmoji(value)) return "名前に絵文字は使用できません";
        break;
      case "email":
        if (!value.trim()) return "メールアドレスは必須です";
        if (value.length > ValidationLimits.EMAIL_MAX_LENGTH)
          return `メールアドレスは${ValidationLimits.EMAIL_MAX_LENGTH}文字以内で入力してください`;
        if (!isValidEmail(value))
          return "有効なメールアドレスを入力してください";
        break;
      case "employee_number":
        if (value && value.length > ValidationLimits.EMPLOYEE_NUMBER_MAX_LENGTH)
          return `社員番号は${ValidationLimits.EMPLOYEE_NUMBER_MAX_LENGTH}文字以内で入力してください`;
        if (value && !isValidEmployeeNumber(value))
          return "社員番号は英数字とハイフンのみ使用できます";
        break;
      case "department":
        if (value && value.length > ValidationLimits.DEPARTMENT_MAX_LENGTH)
          return `所属部署は${ValidationLimits.DEPARTMENT_MAX_LENGTH}文字以内で入力してください`;
        if (value && containsEmoji(value))
          return "所属部署に絵文字は使用できません";
        break;
    }
    return undefined;
  };

  const handleFieldChange = (field: keyof UpdateUserRequest, value: string) => {
    setFormData({ ...formData, [field]: value });

    if (
      field in fieldErrors ||
      field === "name" ||
      field === "email" ||
      field === "employee_number" ||
      field === "department"
    ) {
      const error = validateField(field as keyof FieldErrors, value);
      setFieldErrors((prev) => ({ ...prev, [field]: error }));
    }
  };

  const validateForm = (): boolean => {
    const errors: FieldErrors = {};

    errors.name = validateField("name", formData.name);
    errors.email = validateField("email", formData.email);
    errors.employee_number = validateField(
      "employee_number",
      formData.employee_number || "",
    );
    errors.department = validateField("department", formData.department || "");

    setFieldErrors(errors);
    return !Object.values(errors).some((e) => e !== undefined);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token) return;

    if (!validateForm()) {
      return;
    }

    try {
      setLoading(true);
      setError("");
      await userApi.update(token, user.id, formData);
      onSuccess();
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleInputFocus = (
    e: React.FocusEvent<HTMLInputElement | HTMLSelectElement>,
  ) => {
    e.currentTarget.style.borderColor = "var(--primary)";
    e.currentTarget.style.boxShadow = "0 0 0 3px var(--primary-light)";
  };

  const handleInputBlur = (
    e: React.FocusEvent<HTMLInputElement | HTMLSelectElement>,
    hasError: boolean,
  ) => {
    e.currentTarget.style.borderColor = hasError
      ? "var(--error)"
      : "var(--border)";
    e.currentTarget.style.boxShadow = "none";
  };

  return (
    <div
      className="fixed inset-0 overflow-y-auto h-full w-full flex items-center justify-center z-50"
      style={{ backgroundColor: "rgba(15, 23, 42, 0.6)" }}
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        className="rounded-xl shadow-xl w-96 p-6 border-2 animate-scaleIn"
        style={{
          background: "var(--surface)",
          borderColor: "var(--border)",
        }}
      >
        <h2
          className="text-xl font-bold mb-5"
          style={{ color: "var(--foreground)" }}
        >
          ユーザー情報編集
        </h2>

        {error && (
          <div
            className="px-4 py-3 rounded-lg mb-4 border-2"
            style={{
              background: "var(--error-light)",
              borderColor: "var(--error)",
              color: "var(--error)",
            }}
          >
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label
              className="block text-sm font-semibold mb-2"
              style={{ color: "var(--foreground)" }}
            >
              名前 <span style={{ color: "var(--error)" }}>*</span>
            </label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => handleFieldChange("name", e.target.value)}
              className="w-full px-3 py-2.5 border-2 rounded-lg focus:outline-none transition-all"
              style={{
                background: "var(--surface)",
                borderColor: fieldErrors.name
                  ? "var(--error)"
                  : "var(--border)",
                color: "var(--foreground)",
              }}
              onFocus={handleInputFocus}
              onBlur={(e) => handleInputBlur(e, !!fieldErrors.name)}
              maxLength={ValidationLimits.NAME_MAX_LENGTH}
              required
            />
            {fieldErrors.name && (
              <p className="text-xs mt-1" style={{ color: "var(--error)" }}>
                {fieldErrors.name}
              </p>
            )}
            <p className="text-xs mt-1" style={{ color: "var(--secondary)" }}>
              {ValidationLimits.NAME_MAX_LENGTH}文字以内
            </p>
          </div>

          <div className="mb-4">
            <label
              className="block text-sm font-semibold mb-2"
              style={{ color: "var(--foreground)" }}
            >
              メールアドレス <span style={{ color: "var(--error)" }}>*</span>
            </label>
            <input
              type="email"
              value={formData.email}
              onChange={(e) => handleFieldChange("email", e.target.value)}
              className="w-full px-3 py-2.5 border-2 rounded-lg focus:outline-none transition-all"
              style={{
                background: "var(--surface)",
                borderColor: fieldErrors.email
                  ? "var(--error)"
                  : "var(--border)",
                color: "var(--foreground)",
              }}
              onFocus={handleInputFocus}
              onBlur={(e) => handleInputBlur(e, !!fieldErrors.email)}
              maxLength={ValidationLimits.EMAIL_MAX_LENGTH}
              required
            />
            {fieldErrors.email && (
              <p className="text-xs mt-1" style={{ color: "var(--error)" }}>
                {fieldErrors.email}
              </p>
            )}
          </div>

          <div className="mb-4">
            <label
              className="block text-sm font-semibold mb-2"
              style={{ color: "var(--foreground)" }}
            >
              社員番号
            </label>
            <input
              type="text"
              value={formData.employee_number || ""}
              onChange={(e) =>
                handleFieldChange("employee_number", e.target.value)
              }
              className="w-full px-3 py-2.5 border-2 rounded-lg focus:outline-none transition-all"
              style={{
                background: "var(--surface)",
                borderColor: fieldErrors.employee_number
                  ? "var(--error)"
                  : "var(--border)",
                color: "var(--foreground)",
              }}
              onFocus={handleInputFocus}
              onBlur={(e) => handleInputBlur(e, !!fieldErrors.employee_number)}
              maxLength={ValidationLimits.EMPLOYEE_NUMBER_MAX_LENGTH}
              placeholder="例: EMP-001"
            />
            {fieldErrors.employee_number && (
              <p className="text-xs mt-1" style={{ color: "var(--error)" }}>
                {fieldErrors.employee_number}
              </p>
            )}
            <p className="text-xs mt-1" style={{ color: "var(--secondary)" }}>
              英数字とハイフンのみ、
              {ValidationLimits.EMPLOYEE_NUMBER_MAX_LENGTH}文字以内
            </p>
          </div>

          <div className="mb-4">
            <label
              className="block text-sm font-semibold mb-2"
              style={{ color: "var(--foreground)" }}
            >
              所属部署
            </label>
            <input
              type="text"
              value={formData.department || ""}
              onChange={(e) => handleFieldChange("department", e.target.value)}
              className="w-full px-3 py-2.5 border-2 rounded-lg focus:outline-none transition-all"
              style={{
                background: "var(--surface)",
                borderColor: fieldErrors.department
                  ? "var(--error)"
                  : "var(--border)",
                color: "var(--foreground)",
              }}
              onFocus={handleInputFocus}
              onBlur={(e) => handleInputBlur(e, !!fieldErrors.department)}
              maxLength={ValidationLimits.DEPARTMENT_MAX_LENGTH}
              placeholder="例: 開発部"
            />
            {fieldErrors.department && (
              <p className="text-xs mt-1" style={{ color: "var(--error)" }}>
                {fieldErrors.department}
              </p>
            )}
            <p className="text-xs mt-1" style={{ color: "var(--secondary)" }}>
              {ValidationLimits.DEPARTMENT_MAX_LENGTH}文字以内
            </p>
          </div>

          <div className="mb-6">
            <label
              className="block text-sm font-semibold mb-2"
              style={{ color: "var(--foreground)" }}
            >
              権限 <span style={{ color: "var(--error)" }}>*</span>
            </label>
            <select
              value={formData.role}
              onChange={(e) =>
                setFormData({ ...formData, role: e.target.value as Role })
              }
              className="w-full px-3 py-2.5 border-2 rounded-lg focus:outline-none transition-all"
              style={{
                background: "var(--surface)",
                borderColor: "var(--border)",
                color: "var(--foreground)",
              }}
              onFocus={handleInputFocus}
              onBlur={(e) => handleInputBlur(e, false)}
              required
            >
              <option value="admin">管理者</option>
              <option value="editor">編集者</option>
              <option value="viewer">閲覧者</option>
            </select>
          </div>

          <div className="flex gap-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2.5 border-2 rounded-lg transition-all font-medium"
              style={{
                borderColor: "var(--border)",
                color: "var(--foreground)",
                background: "var(--surface)",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = "var(--secondary-light)";
                e.currentTarget.style.borderColor = "var(--secondary)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = "var(--surface)";
                e.currentTarget.style.borderColor = "var(--border)";
              }}
              disabled={loading}
            >
              キャンセル
            </button>
            <button
              type="submit"
              className="flex-1 px-4 py-2.5 text-white rounded-lg shadow-lg transition-all font-medium disabled:opacity-50"
              style={{ background: "var(--primary)" }}
              onMouseEnter={(e) =>
                !loading &&
                (e.currentTarget.style.background = "var(--primary-hover)")
              }
              onMouseLeave={(e) =>
                (e.currentTarget.style.background = "var(--primary)")
              }
              disabled={loading}
            >
              {loading ? "保存中..." : "保存"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
