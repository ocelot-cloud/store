<template>
  <v-text-field
      :id="id"
      :type="inputType"
      v-model="localValue"
      @input="updateValue"
      :error="submitted && hasError"
      :error-messages="submitted && hasError ? [errorMessageText] : []"
      :placeholder="placeholderText"
      required
  />
</template>

<script lang="ts">
import {defineComponent, computed, watch, ref} from 'vue'
import {
  defaultAllowedSymbols,
  generateInvalidInputMessage,
  getDefaultValidationRegex,
  maxLengthPassword,
  defaultMaxLength,
  minLengthPassword,
  defaultMinLength,
  passwordAllowedSymbols
} from '@/components/shared'

type ValidationType = 'username' | 'password' | 'email' | 'app'

const validationConfig = {
  username: {
    id: 'input-username',
    type: 'text',
    pattern: getDefaultValidationRegex(),
    errorMessage: generateInvalidInputMessage('username', defaultAllowedSymbols, defaultMinLength, defaultMaxLength),
    placeholder: 'Username'
  },
  password: {
    id: 'input-password',
    type: 'password',
    pattern: new RegExp(`^${passwordAllowedSymbols}{${minLengthPassword},${maxLengthPassword}}$`),
    errorMessage: generateInvalidInputMessage('password', passwordAllowedSymbols, minLengthPassword, maxLengthPassword),
    placeholder: 'Password'
  },
  email: {
    id: 'input-email',
    type: 'email',
    pattern: new RegExp(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+.[a-zA-Z]{2,}$`),
    errorMessage: 'Invalid email, must have this format: x@x.x',
    placeholder: 'Email'
  },
  app: {
    id: 'input-app',
    type: 'text',
    pattern: getDefaultValidationRegex(),
    errorMessage: generateInvalidInputMessage('app', defaultAllowedSymbols, defaultMinLength, defaultMaxLength),
    placeholder: 'Enter name of the new app'
  }
}

export default defineComponent({
  name: 'ValidatedInput',
  props: {
    modelValue: {
      type: String,
      required: true
    },
    validationType: {
      type: String as () => ValidationType,
      required: true
    },
    submitted: {
      type: Boolean,
      required: true
    },
    id: {
      type: String,
      required: true
    }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const config = validationConfig[props.validationType as ValidationType]
    const hasError = computed(() => !config.pattern.test(props.modelValue))
    const inputType = computed(() => config.type)
    const errorMessageText = computed(() => config.errorMessage)
    const placeholderText = computed(() => config.placeholder)

    const localValue = ref(props.modelValue)
    watch(localValue, (newValue) => {
      emit('update:modelValue', newValue)
    })

    const updateValue = (event: Event) => {
      emit('update:modelValue', (event.target as HTMLInputElement).value)
    }

    return {
      hasError,
      inputType,
      errorMessageText,
      placeholderText,
      updateValue,
      localValue,
    }
  }
})
</script>
