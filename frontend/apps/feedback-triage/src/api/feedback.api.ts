import { API } from "@moh-sso/config";
import { baseApi } from "@moh-sso/api";

import type { GetFeedbackParams, GetFeedbackResponse, TriageUpdatePayload } from "../types";

type ApiEnvelope<T> = {
  success: boolean;
  data: T;
};

export const feedbackApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getFeedback: builder.query<GetFeedbackResponse, GetFeedbackParams | void>({
      query: (params) => {
        const queryParams = new URLSearchParams();
        if (params) {
          Object.entries(params).forEach(([key, value]) => {
            if (value !== undefined && value !== "") {
              queryParams.append(key, String(value));
            }
          });
        }
        const queryString = queryParams.toString();
        return {
          url: queryString ? `${API.feedback.list()}?${queryString}` : API.feedback.list(),
          method: "GET",
          credentials: "include",
        };
      },
      transformResponse: (res: ApiEnvelope<GetFeedbackResponse>) => res.data,
      providesTags: ["Feedback"],
    }),

    updateFeedbackTriage: builder.mutation<
      { message: string },
      { id: number; body: TriageUpdatePayload }
    >({
      query: ({ id, body }) => ({
        url: API.feedback.triage(id),
        method: "PATCH",
        body,
        credentials: "include",
      }),
      transformResponse: (res: ApiEnvelope<{ message: string }>) => res.data,
      invalidatesTags: ["Feedback"],
    }),
  }),
});

export const { useGetFeedbackQuery, useLazyGetFeedbackQuery, useUpdateFeedbackTriageMutation } =
  feedbackApi;
